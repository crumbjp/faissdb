# Reference

## CLI

```
faissdb [<config-file>] [--fullsync]
```

| Argument | Description |
|---|---|
| `<config-file>` | Path to the YAML config file. Default: `config.yml`. |
| `--fullsync` | Run `FullLocalSync` at startup and exit. See [OPERATIONS.md](OPERATIONS.md#primary-index-corruption-recovery). No gRPC (feature / replica) or HTTP server is started in this mode; the process exits with code 0 on success, non-zero on failure. Argument order with `<config-file>` is not significant. |

## HTTP API

| Method | Path | Description |
|---|---|---|
| `GET` | `/` | Returns server status as JSON (training state, data count, replica set info, ntotal per index). |
| `POST` | `/train` | Train FAISS indexes. Body: proportion as float (0.0 - 1.0). Skipped if already trained. |
| `POST` | `/ftrain` | Force train FAISS indexes. Body: proportion as float. Re-trains even if already trained. |
| `POST` | `/fullsync` | Rebuild FAISS indexes and idDB from dataDB without re-training. Emits `OP_FULLSYNC` oplog. |
| `PUT` | `/replicaset` | Initialize or reset the replica set configuration. Body: JSON. Primary only. |
| `DELETE` | `/replicaset` | Shutdown the replica set. |

### GET / Response

```json
{
  "Status": 1,
  "Istrained": true,
  "Lastsynced": "...",
  "Lastkey": "...",
  "DataCount": 3000000,
  "Faiss": { ... },
  "Ntotal": { "main": 3000000 },
  "ReplicaSet": { ... },
  "Primary": true,
  "Secondary": false
}
```

## gRPC API

Client API is defined in [feature.proto](/protos/feature.proto), replication protocol in [replica.proto](/protos/replica.proto).

### Feature Service (client)

| RPC | Description |
|---|---|
| `Status` | Get server status. |
| `Set` | Upsert a vector with uniqkey and optional collection names. |
| `Del` | Delete a vector by uniqkey. |
| `Search` | Search nearest neighbors by vector. |
| `Train` | Trigger FAISS training. |
| `Dropall` | Drop all data. |
| `DbStats` | Get RocksDB statistics. |

### Replica Service (internal)

| RPC | Description |
|---|---|
| `PrepareResetReplicaSet` | Prepare replica set reconfiguration. |
| `ResetReplicaSet` | Apply replica set reconfiguration. |
| `GetStatus` | Get node status for replication. |
| `GetLastKey` | Get the last oplog key. |
| `GetTrained` | Transfer trained FAISS index data. |
| `GetData` | Transfer all data for full sync. |
| `GetCurrentOplog` | Stream oplog entries for incremental sync. |

# Configuration Reference

faissdb is configured via a YAML file passed as the first argument (default: `config.yml`).

## process

| Key | Type | Description |
|---|---|---|
| `loglv` | string | Log level. One of `debug`, `trace`, `info`, `warn`, `error`, `fatal`. |
| `performancelog` | bool | Enable performance profiling logs. |
| `logfile` | string | Path to the log file. |
| `pidfile` | string | Path to the PID file (used in daemon mode). |
| `daemon` | bool | Run as a daemon process. |
| `memlimit` | int64 | Go runtime memory limit in bytes (`GOMEMLIMIT`). Controls GC aggressiveness. Set to 0 or omit to disable. Should be set lower than the container/host memory limit, but higher than the expected peak Go heap usage. Note: this only affects Go heap memory, not native memory used by FAISS or RocksDB. |

## http

| Key | Type | Description |
|---|---|---|
| `port` | int | HTTP server listen port. |
| `maxconnections` | int | Maximum number of concurrent HTTP connections. |
| `httptimeout` | int | Read/Write timeout in seconds. |

## db

| Key | Type | Description |
|---|---|---|
| `dbpath` | string | Base directory for all data files (RocksDB, FAISS indexes). |

### db.faiss

| Key | Type | Description |
|---|---|---|
| `dimension` | int | Vector dimension. Must match the dimension of vectors being stored. |
| `description` | string | FAISS index factory string (e.g. `IVF1024,Flat`, `IVF256,PQ32x8`). See [FAISS wiki](https://github.com/facebookresearch/faiss/wiki/The-index-factory). |
| `metric` | string | Distance metric. `InnerProduct` or `L2`. |
| `nprobe` | int | Number of clusters to search at query time. Higher values improve recall at the cost of speed. |
| `directmap` | bool | Enable DirectMap (Hashtable) for IVF indexes. When enabled, FAISS maintains an ID-to-inverted-list mapping, allowing O(1) single-vector removal via `remove_ids` without a full scan. Without DirectMap, removal requires scanning all inverted lists. The Hashtable type is used (not Array) to support non-contiguous IDs. Default: `true`. |
| `syncinterval` | int | Interval in milliseconds to periodically write FAISS indexes to disk. |

### db.metadb / db.datadb / db.iddb / db.logdb / db.replicadb

| Key | Type | Description |
|---|---|---|
| `capacity` | uint64 | RocksDB block cache capacity in bytes. |

`db.replicadb` holds the ReplicaSet configuration (members, primary assignment, timestamp). It was split out from `metadb` in 0.3.1 so the cluster configuration can be purged independently of the vector data. On first startup after upgrade, the configuration is auto-migrated from `metadb`.

## oplog

| Key | Type | Description |
|---|---|---|
| `term` | int | Oplog retention period in seconds. Entries older than this are periodically deleted. |

## feature

| Key | Type | Description |
|---|---|---|
| `listen` | string | gRPC listen address for client feature API (e.g. `:20025`). |

## replica

| Key | Type | Description |
|---|---|---|
| `listen` | string | gRPC listen address for replication protocol (e.g. `:21025`). |

## Example

```yaml
process:
  loglv: trace
  performancelog: false
  logfile: /usr/local/faissdb/log/faissdb.log
  pidfile: /usr/local/faissdb/tmp/faissdb.pid
  daemon: false
  memlimit: 12884901888
http:
  port: 9095
  maxconnections: 1000
  httptimeout: 60
db:
  dbpath: /usr/local/faissdb/data
  faiss:
    dimension: 768
    syncinterval: 60000
    description: IVF1024,Flat
    metric: InnerProduct
    nprobe: 32
    directmap: true
  metadb:
    capacity: 1073741824
  datadb:
    capacity: 1073741824
  iddb:
    capacity: 1073741824
  logdb:
    capacity: 1073741824
  replicadb:
    capacity: 1073741824
oplog:
  term: 3600
feature:
  listen: ":20025"
replica:
  listen: ":21025"
```
