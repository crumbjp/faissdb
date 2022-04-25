#!/usr/bin/env bash
CURDIR=`dirname $0`

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

export NODE_ENV=test
export NODE_PATH=src
$CURDIR/node_modules/mocha/bin/mocha --config test/.mocharc.json --exit $@
