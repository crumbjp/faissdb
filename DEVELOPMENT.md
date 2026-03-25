# Development

## Prerequisites

- Docker
- [go-faiss](https://github.com/crumbjp/go-faiss) cloned at `../go-faiss` (sibling of faissdb repository)

## Build

Build the `faissdb:build` Docker image from source.

```
cd docker
./build.sh
```

This creates a base image, compiles faissdb and go-faiss inside a container, then commits the result as `faissdb:build`.

## dev.sh commands

All development operations are done through `docker/dev.sh`.

| Command | Description |
|---|---|
| `start_container` | Create and start the development container from `faissdb:build`. Mounts source, config (`nodejs/example`), data, and log directories. |
| `stop_container` | Stop and remove the development container. |
| `start` | Start the faissdb process inside the container using `nodejs/example/config.yml`. |
| `stop` | Stop the faissdb process (sends kill via PID file). |
| `rebuild` | Recompile the server binary from current source inside the container. |
| `setup` | Initialize a single-node replica set as primary. |
| `train [proportion]` | Train FAISS indexes. `proportion` (default: `1`) controls how much data is used (0.0 - 1.0). |
| `fullsync` | Rebuild FAISS indexes and idDB from dataDB. Emits `OP_FULLSYNC` for secondary nodes. |

### CI / Release commands

| Command | Description |
|---|---|
| `build_ci_container` | Build a CI Docker image. |
| `start_release_container` | Start a container from the release image. |
| `start_manifest_container` | Start a container from the manifest image. |
| `push` | Push the release image to Docker Hub. |
| `manifest` | Create and push a multi-arch manifest. |

## Typical workflow

### First time setup

```
cd docker
./build.sh
./dev.sh start_container
./dev.sh start
./dev.sh setup
```

### Code change and rebuild

```
# Edit source code
cd docker
./dev.sh stop
./dev.sh rebuild
./dev.sh start
```

### Train and sync (after data is loaded)

```
cd docker
./dev.sh train 0.1    # Train with 10% of data (sufficient for IVF)
./dev.sh fullsync     # Rebuild indexes from all data
```

## Config

The development container uses `nodejs/example/config.yml` as its configuration. See [REFERENCE.md](REFERENCE.md) for configuration details.

## Ports (nodejs/example/config.yml)

| Port | Protocol | Description |
|---|---|---|
| 9091 | HTTP | Management API |
| 20021 | gRPC | Feature (client) API |
| 21021 | gRPC | Replica protocol |

## Data and logs

- Data: `docker/build/mnt/data/`
- Logs: `docker/build/mnt/log/faissdb.log`
