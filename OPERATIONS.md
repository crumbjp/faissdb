# Operations

Operational runbooks for faissdb. Aimed at on-call engineers responding to incidents.

## Primary index corruption recovery

### Symptoms

- `GET /` on the primary returns `Ntotal.<collection> = 0` while `DataCount` is non-zero.
- Memory usage (RSS) on the primary is significantly smaller than on secondaries (FAISS indexes are memory-mapped, so a missing index file shrinks RSS).
- The primary serves search requests but returns empty results for the affected collection.
- Startup logs show `FaissIndex[<name>].Open() ReadIndex <error>` followed by a fallback to the trained template (pre-0.3.1 behavior — see below).

### Root cause

FAISS index files are written to disk periodically (`db.faiss.syncinterval`). If the process is killed (OOM, SIGKILL, host crash) during `FaissIndex.Write()`, the on-disk file can be left truncated. On next startup:

- **Up to 0.3.0**: `FaissIndex.Open()` silently overwrote the corrupted file with the empty trained template and continued. The node served traffic with `Ntotal=0`, and only newly written records (via `ReplicaSync` / client writes) were added back. Pre-crash records in `dataDB` remained orphaned.
- **From 0.3.1**: `FaissIndex.Open()` detects a non-empty but unreadable file and the process exits via `Fatal` before serving traffic. Recovery must be performed explicitly as described below.

### Recovery procedure (primary)

Prerequisite: the `dataDB` (RocksDB) is intact and contains all records. This is the normal case — corruption affects only the FAISS index files under `<dbpath>/<collection>`.

0. **Snapshot the current state**
   ```sh
   curl -s http://<primary>:<http-port>/ | jq '{Status,DataCount,Lastkey,Ntotal}'
   curl -s http://<secondary>:<http-port>/ | jq '{Status,DataCount,Lastkey,Ntotal}'
   ```
   Record `DataCount` and `Ntotal` for post-recovery verification.

1. **Back up the data directory** on both nodes (skip only if you accept the risk):
   ```sh
   DBPATH=<db.dbpath from config.yml>
   sudo tar -C "$(dirname "$DBPATH")" \
     -czf /var/backups/faissdb-$(hostname)-$(date +%Y%m%d%H%M).tgz \
     "$(basename "$DBPATH")"
   ```

2. **Stop writes** from the application to faissdb. Reads may continue, but the primary will stop serving (gRPC feature / HTTP) during the rebuild.

3. **Stop the primary service**:
   ```sh
   systemctl stop faissdb   # or your process manager
   ```

4. **Run `--fullsync`** on the primary:
   ```sh
   /usr/local/faissdb/bin/faissdb /usr/local/faissdb/config.yml --fullsync
   ```
   The process rebuilds the FAISS indexes from the local `dataDB` and exits with code 0 on success. Expected log progression:
   ```
   start() --fullsync: running FullLocalSync
   FullLocalSync() start
   LocalIndex.ResetToTrained() start
   LocalIndex.ResetToTrained() Reset index main
   LocalIndex.ResetToTrained() Reset index recent
   LocalIndex.ResetToTrained() Reset index season
   LocalIndex.ResetToTrained() end
   FullLocalSync() ResetToTrained done
   FullLocalSync() idDB reopened
   LocalIndex.SyncFromLocalDb() start
   LocalIndex.SyncFromLocalDb() synced 10000
   ...
   LocalIndex.SyncFromLocalDb() sync complete: <DataCount> records
   LocalIndex.Write() end <lastkey>
   FullLocalSync() end
   start() --fullsync: done, shutting down
   ```
   During this mode neither the feature gRPC, replica gRPC, nor the HTTP server is started. The node is fully isolated from the cluster.

5. **Restart the primary normally**:
   ```sh
   systemctl start faissdb
   ```

6. **Verify**:
   ```sh
   curl -s http://<primary>:<http-port>/ | jq '{Status,DataCount,Lastkey,Ntotal}'
   ```
   - `Status: 100` (`STATUS_READY`)
   - `Ntotal.<collection>` ≈ `DataCount`
   - RSS has grown to the expected size (approximately `DataCount × dimension × 4 bytes` plus overhead)

7. **Secondary cascade resync (automatic)**
   `FullLocalSync` writes an `OP_FULLSYNC` oplog entry at the end. When the primary comes back online, each secondary connects via replica gRPC, detects the `OP_FULLSYNC` entry, and triggers `ReplicaFullSync` on itself. Monitor secondary logs:
   ```sh
   tail -F /var/log/faissdb.log | grep -E 'ReplicaFullSync|OP_FULLSYNC'
   ```
   Secondaries will briefly go through `STATUS_FULLSYNC` before returning to `STATUS_READY`.

   If you need to avoid the secondary cascade (e.g. maintenance window constraints), stop the secondary before restarting the primary, and only start it once you are ready to absorb the cascade.

8. **Resume application writes.**

### Recovery procedure (secondary)

A secondary with a corrupted FAISS index is recovered by removing its data directory and letting the replica-set bootstrap re-sync from the primary. `--fullsync` can be used on a secondary, but it is not the recommended path because it rebuilds only from its own local `dataDB` (which may itself be stale); re-bootstrapping from the primary is simpler and guarantees consistency.

```sh
systemctl stop faissdb
rm -rf "$DBPATH"/*
systemctl start faissdb
```

## Upgrading 0.3.x to 0.4.x

Verified end-to-end by [ci/test_migration.sh](ci/test_migration.sh) (build a 0.3.1 cluster, upgrade the secondary then the primary in place on the same data, roll both back).

### Compatibility summary

| Layer | Compatibility |
|---|---|
| Replica protocol (gRPC) | Unchanged. Mixed clusters work in both directions: a 0.4.0 node applies old-format oplog with the pre-0.4.0 force semantics; a 0.3.x node ignores the new delta fields and force-applies. Upgrade order is free; simultaneous upgrade is NOT required. |
| Feature protocol (clients) | Existing rpcs unchanged; old clients work against 0.4.0. `SetCollections` returns `UNIMPLEMENTED` on pre-0.4.0 servers — start using it only after every node runs 0.4.0. |
| dataDB / idDB / metaDB / oplogDB | No format change. |
| FAISS index / trained files | No format change (faiss 1.14.1 on both sides). |
| oplog | 0.4.0 reads 0.3.x entries (gap replay included), and 0.3.x reads 0.4.0 delta-carrying entries (rollback safe). |

### Procedure (per node, secondaries first, primary last)

1. **Wait for the index flush before stopping a 0.3.x node.** 0.3.x cannot flush its FAISS indexes on shutdown (its shutdown path exits before the flush; fixed in 0.4.1), so the on-disk indexes are only as fresh as the last periodic sync (`db.faiss.syncinterval`). Poll `DbStats` until `lastsynced` reaches the `lastkey` value observed after your last write. This matters most right after a full sync or `/train`, whose base data is not replayable from the local oplog.
2. Stop the node **with SIGTERM**. Note the docker images up to 0.4.0 declare `STOPSIGNAL SIGRTMIN+3`, which the server does not handle — a plain `docker stop` kills the process abruptly. Use `docker kill -s TERM <container> && docker wait <container>` instead (images 0.4.1 and later declare SIGTERM and `docker stop` becomes safe).
3. Start the 0.4.0 binary/image on the same data directory. Gap sync replays the oplog tail into the indexes automatically.
4. Verify with `GET /` or `DbStats` (counts, `lastsynced` advancing), then move to the next node.
5. Primary last: stopping the primary is the only write-downtime window (there is no automatic failover). Reads keep being served by secondaries.

If a 0.3.x node was killed abruptly right after a full sync (empty indexes but populated dataDB after restart — `Ntotal` near 0 while `DataCount` is large), rebuild locally with `faissdb <config.yml> --fullsync` before serving.

### Rollback

0.4.0 → 0.3.x on the same data directory is supported: 0.3.x ignores the delta fields in 0.4.0-written oplog entries and force-applies them. Stop the 0.4.0 node with SIGTERM (its shutdown flush works), start the 0.3.x binary, and stop using `SetCollections` before rolling back the primary.

## Purging ReplicaSet configuration

When the cluster configuration gets into a bad state (e.g. inconsistent `rsTs` across nodes, wrong primary designation, orphan member entries) and you want to reset the ReplicaSet without touching the vector data or oplog, remove only `replicadb`:

```sh
systemctl stop faissdb
DBPATH=<db.dbpath from config.yml>
sudo rm -rf "$DBPATH"/replica
systemctl start faissdb
```

On restart the node comes up with ReplicaSet unconfigured (`selfMember == nil`). Then re-issue `PUT /replicaset` to the intended primary to re-form the cluster. `dataDB`, `idDB`, `metaDB`, `oplogDB`, and the FAISS index files are untouched.

This separation was introduced in 0.3.1; before that `ReplicaSetTs` / `ReplicaSet` were stored in `metaDB` alongside `lastkey` / collection metadata, so purging them required removing the entire data directory.

## Diagnostics

### Confirm which index file is corrupted

Compare file sizes on primary vs. secondary:

```sh
ls -la "$DBPATH"/{main,recent,season,faiss_trained}
stat "$DBPATH/main"
```

A primary-side index file that is dramatically smaller than the secondary's equivalent is a strong corruption signal.

### Sanity check after startup

```sh
DC=$(curl -s http://<node>:<http-port>/ | jq '.DataCount')
NT=$(curl -s http://<node>:<http-port>/ | jq '.Ntotal')
echo "DataCount=$DC  Ntotal=$NT"
```

If `DataCount > 0` and any `Ntotal.*` is `0` (while that collection is supposed to hold data), treat as incident. From 0.3.1 the process should have failed to start in this state; if it did not, the corruption is in a different component.

### Check for OOM at the time of the previous shutdown

```sh
dmesg -T | grep -i -E 'kill|oom' | tail
journalctl -k | grep -i oom | tail
journalctl -u faissdb --since "<time>" --until "<time>"
```

An OOM kill during `FaissIndex[...].Write()` is the most common trigger for index file truncation.
