##
```
apt update
apt install build-essential
apt install git vim wget libssl-dev
```

### /etc/bash.bashrc
```
export GOENV_ROOT=/usr/local/goenv
export PATH=$GOENV_ROOT/bin:$PATH
eval "$(goenv init -)"
export PATH=$PATH:$GOPATH/bin
```

## golang
```
git clone https://github.com/syndbg/goenv.git /usr/local/goenv
export GOENV_ROOT=/usr/local/goenv
export PATH=$GOENV_ROOT/bin:$PATH
goenv install 1.15.9
goenv global 1.15.9
eval "$(goenv init -)"
export PATH=$PATH:$GOPATH/bin
```

### cmake
```
wget https://github.com/Kitware/CMake/releases/download/v3.20.0/cmake-3.20.0.tar.gz
tar xzf cmake-3.20.0.tar.gz
cd cmake-3.20.0
./bootstrap
make
make install
```

### rockdb
```
apt install libsnappy-dev liblz4-dev libzstd-dev
git clone https://github.com/facebook/rocksdb.git
cd rocksdb
git checkout -b v6.15.5 refs/tags/v6.15.5
make release
make release install

```

### faiss
```
apt install libopenblas-base libopenblas-dev
git clone https://github.com/facebookresearch/faiss.git
cd faiss
git checkout -b 4f12d9c20ca34f7c880b30fa51db9d32cbb47f87 4f12d9c20ca34f7c880b30fa51db9d32cbb47f87
cmake -B build -DFAISS_ENABLE_GPU=OFF -DFAISS_ENABLE_PYTHON=OFF -DFAISS_ENABLE_C_API=ON -DBUILD_SHARED_LIBS=ON .
make -C build
make -C build install
```

#### To fix bug
```
cp build/c_api/libfaiss_c.so  /usr/local/lib
mv /usr/local/include/faiss/c_api/c_api/*.h /usr/local/include/faiss/c_api/
mv /usr/local/include/faiss/c_api/c_api/impl/AuxIndexStructures_c.h  /usr/local/include/faiss/c_api/impl/
```

## protobuf
```
apt install zip
wget https://github.com/protocolbuffers/protobuf/releases/download/v3.15.8/protoc-3.15.8-linux-x86_64.zip
unzip protoc-3.15.8-linux-x86_64.zip
cp bin/protoc /usr/local/bin/
cp -r include/google /usr/local/include/
```

## grpc
```
export GO111MODULE=on
go get google.golang.org/protobuf/cmd/protoc-gen-go google.golang.org/grpc/cmd/protoc-gen-go-grpc
```

### faissdb
apt install libbz2-dev zlib1g-dev
git clone https://github.com/crumbjp/faissdb.git
cd faissdb
git submodule init
git submodule update
cd server
make

### edit `go env GOPATH`/pkg/mod/github.com/tecbot/gorocksdb@[???]/dynflag.go
from: // #cgo LDFLAGS: -lrocksdb -lstdc++ -lm -ldl
to: // #cgo LDFLAGS: -L/usr/local/lib -lrocksdb -lstdc++ -lm -lz -lbz2 -lsnappy -llz4 -lzstd

make
