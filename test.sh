#!/usr/bin/env bash
set -eu

pushd `dirname $0`/server
go test -v
popd
