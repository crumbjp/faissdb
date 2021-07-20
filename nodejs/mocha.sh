#!/usr/bin/env bash
CURDIR=`dirname $0`

export NODE_ENV=test
export NODE_PATH=src
$CURDIR/node_modules/mocha/bin/mocha --config test/.mocharc.json --exit $@
