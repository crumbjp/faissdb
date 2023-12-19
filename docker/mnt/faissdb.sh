#!/usr/bin/env bash
cd /mnt

cp -r /mnt/goenv /usr/local/
export GOENV_ROOT=/usr/local/goenv
export PATH=$GOENV_ROOT/bin:$PATH
export GO111MODULE=on
eval "$(goenv init -)"

echo '
export GOENV_ROOT=/usr/local/goenv
export PATH=$GOENV_ROOT/bin:$PATH
eval "$(goenv init -)"
export GO111MODULE=on
' >> /etc/profile

goenv install 1.15.9
goenv global 1.15.9

cd /mnt/rocksdb
make static_lib
make install

cd /mnt/faiss/faiss-1.7.3
cmake -B build -DFAISS_ENABLE_GPU=OFF -DFAISS_ENABLE_PYTHON=OFF -DFAISS_ENABLE_C_API=ON -DBUILD_SHARED_LIBS=ON -DCMAKE_BUILD_TYPE=Release .
make -C build install
cp build/c_api/libfaiss_c.so  /usr/local/lib
ldconfig

cd /mnt/protoc
cp bin/protoc /usr/local/bin/
cp -r include/google /usr/local/include/

go get google.golang.org/protobuf/cmd/protoc-gen-go google.golang.org/grpc/cmd/protoc-gen-go-grpc

cd /mnt/faissdb
git submodule init
git submodule update
cd server
make && :
GOPATH=`go env GOPATH`
DYNFLAG_GO=`find $GOPATH/pkg/mod/github.com/tecbot -name dynflag.go`
echo '
// +build !linux !static

package gorocksdb

// #cgo LDFLAGS: -L/usr/local/lib -lrocksdb -lstdc++ -lm -lz -lbz2 -lsnappy -llz4 -lzstd -ldl
import "C"
' > $DYNFLAG_GO
echo 'Re build with fixed LDFLAGS'
make

mkdir -p /usr/local/faissdb/bin /usr/local/faissdb/data /usr/local/faissdb/log /usr/local/faissdb/tmp
cp -f server/faissdb /usr/local/faissdb/bin/faissdb
