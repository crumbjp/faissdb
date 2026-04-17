# CHANGELOG

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
