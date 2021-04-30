#!/usr/bin/env bash
set -eu

rm -rf /tmp/faissdb_primary /tmp/faissdb_secondary
mkdir -p /tmp/faissdb_primary/log
mkdir -p /tmp/faissdb_primary/tmp
mkdir -p /tmp/faissdb_primary/data
mkdir -p /tmp/faissdb_secondary/log
mkdir -p /tmp/faissdb_secondary/tmp
mkdir -p /tmp/faissdb_secondary/data

pushd `dirname $0`/server
go test -v
popd
