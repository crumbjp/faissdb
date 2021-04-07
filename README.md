## OSX
### Setup GO
#### goenv
```
git clone https://github.com/syndbg/goenv.git ~/.goenv

```

#### .bashrc
```
export GOENV_ROOT=$HOME/.goenv
export PATH=$GOENV_ROOT/bin:$PATH
eval "$(goenv init -)"
export PATH=$PATH:$GOPATH/bin
```

#### install
```
goenv install 1.15.9
```

### faiss
```
git clone https://github.com/facebookresearch/faiss.git
cd faiss
cmake -B build -DFAISS_ENABLE_GPU=OFF -DFAISS_ENABLE_PYTHON=OFF -DFAISS_ENABLE_C_API=ON -DBUILD_SHARED_LIBS=ON .
make -C build
make -C build install
cp build/c_api/libfaiss_c.dylib  /usr/local/lib
```

#### To fix bug
```
mv /usr/local/include/faiss/c_api/c_api/*.h /usr/local/include/faiss/c_api/
mv /usr/local/include/faiss/c_api/c_api/impl/AuxIndexStructures_c.h  /usr/local/include/faiss/c_api/impl/
```


### rocksdb
```
brew install rocksdb
```

### go get
```
go get github.com/DataIntelligenceCrew/go-faiss
CGO_CFLAGS="-I/usr/local/include" \
CGO_LDFLAGS="-L/usr/local/lib -lrocksdb -lstdc++ -lm -lz -lbz2 -lsnappy -llz4 -lzstd" \
go get github.com/tecbot/gorocksdb


//go get github.com/danielmorsing/rocksdb

```


git clone https://github.com/mania25/faiss-go-wrapper.git
cd faiss-go-wrapper/faiss
make
`/usr/local/lib/libfaiss-wrapper.a`
