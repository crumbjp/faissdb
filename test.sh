#!/usr/bin/env bash
set -eu

rm -rf /tmp/faissdb1 /tmp/faissdb2 /tmp/faissdb3
mkdir -p /tmp/faissdb1/log
mkdir -p /tmp/faissdb1/tmp
mkdir -p /tmp/faissdb1/data
mkdir -p /tmp/faissdb2/log
mkdir -p /tmp/faissdb2/tmp
mkdir -p /tmp/faissdb2/data
mkdir -p /tmp/faissdb3/log
mkdir -p /tmp/faissdb3/tmp
mkdir -p /tmp/faissdb3/data

echo $HOME
if [ "${UID}" = "0" ]; then
    if [ "${HOME}" != "/root" ]; then
        export GOPATH=/root/go
    fi
fi

echo $GOPATH
ls -la  $GOPATH

pushd `dirname $0`/server
go test -v
popd
