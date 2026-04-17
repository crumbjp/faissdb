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
