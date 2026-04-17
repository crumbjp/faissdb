#!/usr/bin/env bash
# CI orchestrator. Assumes this repo is checked out as "faissdb" and
# crumbjp/go-faiss is checked out as a sibling directory "go-faiss",
# both under the same parent.
set -eux
cd "$(dirname "$0")/.."

# Build faissdb:build image (and keep the build container up so tests can
# exec into it). This installs Go (per server/.go-version), RocksDB, FAISS,
# protoc, and compiles the faissdb server binary.
./docker/build.sh keep

# Run server unit tests inside the build container.
docker exec faissdb-build bash --login /mnt/faissdb/ci/test_faissdb.sh

# Run nodejs client tests inside the build container. nodenv + node are
# installed on demand from within the script.
docker exec faissdb-build bash --login /mnt/faissdb/ci/test_client.sh

# Cleanup
docker rm -f faissdb-build
