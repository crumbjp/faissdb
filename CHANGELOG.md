# CHANGELOG

## 0.4.0

### New Features
- **`SetCollections` API**: New Feature rpc to update collection membership of an existing key without resending the vector. The server looks up the stored record in `dataDB`, inherits its vector, and applies only the collection diff to the FAISS indexes. Unknown keys are counted as errors. nodejs client: `Client.setCollections` / `ReplicaSet.setCollections`.

### Improvements
- **Minimal FAISS dispatch on `Set`**: `Set` now compares the incoming record with the stored record in `dataDB`. If the vector is unchanged, only collections whose membership changed receive FAISS add/remove; collections the record stays in are not touched. Unchanged vector + unchanged collections performs no FAISS operation at all.
- **Delta-carrying oplog**: `OP_SET` oplog entries now embed the removed/added collections computed by the primary (new `FaissdbRecord.delta` / `removed_collections` / `added_collections` fields). Secondary tailing and gap-sync replay apply exactly that delta when the local record matches the delta's pre-image collections, and fall back to the previous unconditional remove+add on mismatch (crash re-application against final state, divergence). Old-format oplog entries and `ReplicaFullSync` keep the previous force semantics, so rolling upgrades are safe in both directions (old nodes ignore the new fields).
- **Release image rebuilt `FROM scratch`: 829MB → 56MB**: the final image contains only the runtime dependency closure — the stripped server binary (Go runtime statically linked in), stripped `libfaiss` / `libfaiss_c` / `librocksdb` (314MB → 11MB), and the shared libraries resolved transitively via `ldd` (glibc loader chain, libstdc++, libgomp, OpenBLAS + libgfortran, compression libs, gflags) plus `nsswitch.conf` and a prebuilt `ld.so.cache`. No shell, no package manager, no Go toolchain (previously goenv + a full Go install were baked in unconditionally; only the CI image needs them). The image has no shell, so `docker exec` debugging is unavailable and runtime directories must be provided as volume mounts.
- **Release-image E2E harness (`ci/test_release.sh`)**: verifies the release image as shipped. Three cluster nodes run only the bundled binary; the full mocha suite runs in a separate CI-image driver container and connects over the network (nodes join the driver's network namespace, so the suite's localhost endpoints work unchanged). Cluster nodes are started on demand through the docker API to preserve the suite's node-join scenario. The suite's server-spawn command is now overridable via the `FAISSDB` env var.

## 0.3.1

### New Features
- **`--fullsync` CLI flag**: Run `faissdb <config.yml> --fullsync` to execute `FullLocalSync` at startup and exit. No gRPC or HTTP servers are started during this mode, so the node is fully isolated while rebuilding FAISS indexes from the local `dataDB`. Intended for recovering a primary whose on-disk FAISS index files were lost or corrupted (e.g. truncated by OOM during `Write`). See [OPERATIONS.md](OPERATIONS.md#primary-index-corruption-recovery).
- **Separate `replicadb` for ReplicaSet configuration**: `ReplicaSetTs` / `ReplicaSet` are now stored in a dedicated RocksDB instance under `<dbpath>/replica` instead of mixed into `metaDB`. Purging the cluster configuration (e.g. recovering from a split-brain or re-forming the ReplicaSet from scratch) is now `rm -rf <dbpath>/replica` without touching vector data. Requires adding `db.replicadb.capacity` to `config.yml`. On first startup after upgrade the existing `ReplicaSetTs` / `ReplicaSet` entries are auto-migrated from `metaDB` to `replicaDB`. See [OPERATIONS.md](OPERATIONS.md#purging-replicaset-configuration).

### Improvements
- **Corrupted index detection (fail-fast)**: `FaissIndex.Open()` now distinguishes between a missing index file (legitimate bootstrap path) and an existing-but-unreadable file (corruption). In the latter case, `OpenAllIndex` now logs `Fatal` and exits instead of silently overwriting the file with the empty trained template, which previously caused the node to keep running with `Ntotal=0` while raw data remained in `dataDB`.

### Bug Fixes
- **LocalDB resource release order**: `LocalDB.DestroyDb` and `LocalDB.Close` now release RocksDB resources in reverse order of construction (WriteOptions / ReadOptions → DB → Options → BlockBasedTableOptions). The previous order destroyed `BlockBasedTableOptions` (which owns the LRU block cache) before closing the DB, leaving the block cache's native allocation unreleased after `DestroyDb`. This manifested during `FullLocalSync` as an extra ~1GB of native RSS per configured `iddb.capacity` carried into the subsequent `SyncFromLocalDb` phase, contributing to OOM on tight memory budgets.
- **Atomic FAISS index writes**: `FaissIndex.flush` now writes to `<path>.tmp`, fsyncs the fd (via `go-faiss` `WriteIndexFsync`), and renames to the final path. `FaissIndex.Open` removes any leftover `.tmp` at startup. Previously `faiss.WriteIndex` opened the final path directly and truncated it to zero before streaming new content, so any process kill (SIGKILL, systemctl stop timeout, OOM Killer) during write left a corrupted partial file on disk — surfacing as `read error: N != M` on next startup. Requires go-faiss with `WriteIndexFsync` (commit `8bb765c` or later).
- **`LocalIndex.SyncFromLocalDb` log format**: removed stray `%s` + `start` (copy-paste leftover from `SyncLocalOplog`) that referenced the package-level `start()` function. `go vet` (run by `go test`) flagged it as "format %s arg start is a func value, not called" and blocked CI.

## 0.3.0

### New Features
- **DirectMap support**: Added `directmap` config option for FAISS IVF indexes. Uses `DirectMap::Hashtable` to enable efficient single-vector removal via `RemoveIDsArray`.
- **`OP_FULLSYNC` oplog**: New oplog type (`OP_FULLSYNC = 4`) emitted after `FullLocalSync`. Secondary nodes receiving this oplog trigger a full resync (`ReplicaFullSync`) instead of incremental replay.
- **`memlimit` config**: Added `process.memlimit` configuration to set Go runtime's `GOMEMLIMIT` via `debug.SetMemoryLimit()`. Helps control GC behavior for large heap workloads.
- **HTTP `/fullsync` endpoint**: Trigger `FullLocalSync` via `POST /fullsync`.
- **Memory diagnostics**: All major log lines now include `alloc`, `sys`, and `rss` (from `/proc/self/statm`) to track both Go heap and native (FAISS/RocksDB) memory usage.

### Improvements
- **`SyncRaw` for FullLocalSync**: Introduced `SyncRaw` function that only rebuilds `idDB` and FAISS indexes during `SyncFromLocalDb`, skipping redundant `dataDB.Put` and per-record oplog generation. Significantly reduces memory usage and I/O during sync.
- **Memory leak fixes in `buildTrainData`**: Fixed `defer key.Free()` / `defer value.Free()` inside loops that prevented memory release until function return. Now freed immediately per iteration.
- **`_TRAIN_` index cleanup**: Prevented the temporary `_TRAIN_` index from being registered in `metaDB`, avoiding unnecessary persistence and Write operations.

### Dependencies
- go-faiss: Updated with `SetDirectMapType` and `RemoveIDsArray` support.

## 0.2.0
- Initial public release.
