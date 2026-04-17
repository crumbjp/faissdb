#!/usr/bin/env bash
# Runs server unit tests inside the faissdb-build container.
# Assumes /etc/profile (sourced by `bash --login`) already sets up goenv
# with the Go version pinned by server/.go-version.
set -eux

rm -rf /tmp/faissdb1 /tmp/faissdb2 /tmp/faissdb3 2>/dev/null || true
mkdir -p /tmp/faissdb1/log /tmp/faissdb1/tmp /tmp/faissdb1/data
mkdir -p /tmp/faissdb2/log /tmp/faissdb2/tmp /tmp/faissdb2/data
mkdir -p /tmp/faissdb3/log /tmp/faissdb3/tmp /tmp/faissdb3/data

cd /mnt/faissdb/server
go test -v
