# CHANGELOG

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
