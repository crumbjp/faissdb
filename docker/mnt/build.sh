#!/usr/bin/env bash
cd `dirname $0`

git clone https://github.com/syndbg/goenv.git

git clone https://github.com/facebook/rocksdb.git
pushd rocksdb
git checkout -b v6.15.5 refs/tags/v6.15.5
popd

rm -rf faiss
mkdir faiss
pushd faiss
wget https://github.com/facebookresearch/faiss/archive/refs/tags/v1.7.3.tar.gz
tar xzvf v1.7.3.tar.gz
popd

rm -rf protoc
mkdir protoc
pushd protoc
wget https://github.com/protocolbuffers/protobuf/releases/download/v3.15.8/protoc-3.15.8-linux-x86_64.zip
unzip protoc-3.15.8-linux-x86_64.zip
popd

rm -rf faissdb
git clone https://github.com/crumbjp/faissdb.git

docker rm faissdb-base
docker rmi crumbjp/faissdb:base
docker rmi crumbjp/faissdb

docker % docker image build -t crumbjp/faissdb:base .
docker run --name=faissdb-base -ti --tmpfs /run --tmpfs /run/lock --tmpfs /tmp:exec \
 --cap-add SYS_ADMIN --device=/dev/fuse \
 --security-opt apparmor:unconfined \
 --security-opt seccomp:unconfined \
 -v /sys/fs/cgroup:/sys/fs/cgroup:ro \
 -v /lib/modules:/lib/modules:ro \
 -v `pwd`/mnt:/mnt \
 -d crumbjp/faissdb:base \
 /bin/bash

docker exec faissdb-base bash --login /mnt/faissdb.sh

docker commit faissdb-base crumbjp/faissdb
