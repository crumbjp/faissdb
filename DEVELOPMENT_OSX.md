## golang
```
git clone https://github.com/syndbg/goenv.git ~/.goenv

```

### .bashrc
```
export GOENV_ROOT=$HOME/.goenv
export PATH=$GOENV_ROOT/bin:$PATH
eval "$(goenv init -)"
export PATH=$PATH:$GOPATH/bin
```

### install
```
goenv install 1.15.9
goenv global 1.15.9
eval "$(goenv init -)"
export PATH=$PATH:$GOPATH/bin
```

## brew
```
brew install cmake
brew install libomp

```

## faiss
```
git clone https://github.com/facebookresearch/faiss.git
cd faiss
git checkout -b v1.7.0 refs/tags/v1.7.0
cmake -B build -DFAISS_ENABLE_GPU=OFF -DFAISS_ENABLE_PYTHON=OFF -DFAISS_ENABLE_C_API=ON -DBUILD_SHARED_LIBS=ON .
make -C build
make -C build install
```

### To fix bug
```
cp build/c_api/libfaiss_c.dylib  /usr/local/lib
mv /usr/local/include/faiss/c_api/c_api/*.h /usr/local/include/faiss/c_api/
mv /usr/local/include/faiss/c_api/c_api/impl/AuxIndexStructures_c.h  /usr/local/include/faiss/c_api/impl/
```

## rocksdb
```
git clone https://github.com/facebook/rocksdb.git
cd rocksdb
git checkout -b v6.15.5 refs/tags/v6.15.5
make
make install
```

## protobuf
```
wget https://github.com/protocolbuffers/protobuf/releases/download/v3.15.8/protoc-3.15.8-osx-x86_64.zip
unzip protoc-3.15.8-osx-x86_64.zip
cp bin/protoc /usr/local/bin/
cp -r include/google /usr/local/include/
```

## grpc
```
export GO111MODULE=on
go get google.golang.org/protobuf/cmd/protoc-gen-go google.golang.org/grpc/cmd/protoc-gen-go-grpc
```

### faissdb
brew install protobuf
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
