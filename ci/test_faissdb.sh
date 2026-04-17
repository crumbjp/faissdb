#!/usr/bin/env bash
# Runs server unit tests. Designed to run inside the CI image
# (crumbjp/faissdb:<ver>-ci), which provides goenv + Go, nodenv + Node,
# pre-built RocksDB / FAISS / protoc, and /usr/local/.faissdb for env
# initialization.
#
# Expected layout: this repo checked out at ${faissdb_root}, and
# crumbjp/go-faiss at ${faissdb_root}/../go-faiss (sibling), because
# server/go.mod has `replace github.com/crumbjp/go-faiss => ../../go-faiss`.
set -eux

rm -rf /tmp/faissdb1 /tmp/faissdb2 /tmp/faissdb3 2>/dev/null || true
mkdir -p /tmp/faissdb1/log /tmp/faissdb1/tmp /tmp/faissdb1/data
mkdir -p /tmp/faissdb2/log /tmp/faissdb2/tmp /tmp/faissdb2/data
mkdir -p /tmp/faissdb3/log /tmp/faissdb3/tmp /tmp/faissdb3/data

. /usr/local/.faissdb

cd "$(dirname "$0")/../server"
go mod tidy
go test -v
