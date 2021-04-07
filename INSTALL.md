## Setup GO
### goenv
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
```

## faiss
### Ubuntu18.04(arm)
#### cmake
```
wget https://github.com/Kitware/CMake/releases/download/v3.20.0/cmake-3.20.0.tar.gz
tar xzf cmake-3.20.0.tar.gz
cd cmake-3.20.0
make
make install
```

#### BLAS
```
apt install libopenblas-base libopenblas-dev
```

### faiss
```
git clone https://github.com/facebookresearch/faiss.git
cd faiss
/usr/local/bin/cmake -B build -DFAISS_ENABLE_GPU=OFF -DFAISS_ENABLE_PYTHON=OFF -DFAISS_ENABLE_C_API=ON -DBUILD_SHARED_LIBS=ON .
make
make install
cp c_api/libfaiss_c.so /usr/local/lib/
```

### OSX
```
git clone https://github.com/facebookresearch/faiss.git
cd faiss
cmake -B build -DFAISS_ENABLE_GPU=OFF -DFAISS_ENABLE_PYTHON=OFF -DFAISS_ENABLE_C_API=ON -DBUILD_SHARED_LIBS=ON .
git clone https://github.com/facebookresearch/faiss.git
make -C build
make -C build install
cp build/c_api/libfaiss_c.dylib  /usr/local/lib
```

### To fix bug
```
mv /usr/local/include/faiss/c_api/c_api/*.h /usr/local/include/faiss/c_api/
mv /usr/local/include/faiss/c_api/c_api/impl/AuxIndexStructures_c.h  /usr/local/include/faiss/c_api/impl/
```


## rocksdb
### OSX
```
brew install rocksdb
```

### Ubuntu18.04(arm)
```
apt install libsnappy-dev liblz4-dev libzstd-dev

git clone https://github.com/facebook/rocksdb.git
cd rocksdb
git reset v6.15.5
make all
make install
```

## go get
### Ubuntu18.04(arm)
```
go get github.com/DataIntelligenceCrew/go-faiss
go get github.com/tecbot/gorocksdb

Edit `go env GOPATH`/pkg/mod/github.com/tecbot/gorocksdb@[???]/dynflag.go
LDFLAGS like bellow

// #cgo LDFLAGS: -L/usr/local/lib -lrocksdb -lstdc++ -lm -lz -lbz2 -lsnappy -llz4 -lzstd

```

### OSX
```
go get github.com/DataIntelligenceCrew/go-faiss
CGO_CFLAGS="-I/usr/local/include" \
CGO_LDFLAGS="-L/usr/local/lib -lrocksdb -lstdc++ -lm -lz -lbz2 -lsnappy -llz4 -lzstd" \
go get github.com/tecbot/gorocksdb
```

## Run faissdb
```
git clone https://github.com/crumbjp/faissdb.git
cd faissdb
LD_LIBRARY_PATH=/usr/local/lib:$LD_LIBRARY_PATH go run server.go v

```
