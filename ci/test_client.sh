#!/usr/bin/env bash
# Runs the nodejs client mocha suite. Designed to run inside the CI image
# (crumbjp/faissdb:<ver>-ci). Builds a fresh faissdb binary from the
# current sources first because mocha launches the server process.
set -eux

rm -rf /tmp/faissdb1 /tmp/faissdb2 /tmp/faissdb3 2>/dev/null || true
mkdir -p /tmp/faissdb1/log /tmp/faissdb1/tmp /tmp/faissdb1/data
mkdir -p /tmp/faissdb2/log /tmp/faissdb2/tmp /tmp/faissdb2/data
mkdir -p /tmp/faissdb3/log /tmp/faissdb3/tmp /tmp/faissdb3/data

. /usr/local/.faissdb

cd "$(dirname "$0")/../server"
go mod tidy
make
cp -f faissdb /usr/local/faissdb/bin/faissdb

cd ../nodejs
npm install
bash mocha.sh
