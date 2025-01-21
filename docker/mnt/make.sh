#!/usr/bin/env bash
rm -rf /mnt/faissdb-build
mkdir /mnt/faissdb-build
cp -a /mnt/faissdb/protos /mnt/faissdb-build
cp -a /mnt/faissdb/server /mnt/faissdb-build
chown root:root -R /mnt/faissdb-build
cd /mnt/faissdb-build/server
go mod tidy
make
cp -f faissdb /usr/local/faissdb/bin/faissdb
