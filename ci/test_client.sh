#!/usr/bin/env bash
set -eu

rm -rf /tmp/faissdb1/data/*
rm -rf /tmp/faissdb2/data/*
rm -rf /tmp/faissdb3/data/*

pushd `dirname $0`/../nodejs
bash mocha.sh
popd
