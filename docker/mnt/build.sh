#!/usr/bin/env bash


if [ "$1" == "release" ]; then
  echo Do nothing
else
  mkdir -p /usr/local
  rmdir /usr/local/include/
  rmdir /usr/local/lib/
  rmdir /usr/local/bin/
  mkdir -p /mnt/local/bin
  mkdir -p /mnt/local/lib
  mkdir -p /mnt/local/include
  ln -sT /mnt/local/lib /usr/local/lib
  ln -sT /mnt/local/include /usr/local/include
  ln -sT /mnt/local/bin /usr/local/bin
  ls -la /usr/local/
  ls -la /usr/local/include/
  ls -la /usr/local/lib/
  ls -la /usr/local/bin/
fi

cd /mnt

if [ ! -d /mnt/goenv ]; then
  git clone https://github.com/syndbg/goenv.git /mnt/goenv
fi
cp -r /mnt/goenv /usr/local/
export GOENV_ROOT=/usr/local/goenv
export PATH=$GOENV_ROOT/bin:$PATH
export GO111MODULE=on
eval "$(goenv init -)"

echo '
export GOENV_ROOT=/usr/local/goenv
export PATH=$GOENV_ROOT/bin:$PATH
export GO111MODULE=on
eval "$(goenv init -)"
' >> /etc/profile

goenv install 1.23.4
goenv global 1.23.4

if [ "$1" != "release" ]; then
  if [ -d /mnt/rocksdb ]; then
    cd /mnt/rocksdb
    export PORTABLE=1
    make install-shared
  else
    git clone https://github.com/facebook/rocksdb.git /mnt/rocksdb
    cd /mnt/rocksdb
    git checkout -b v9.8.4 refs/tags/v9.8.4
    export PORTABLE=1
    make shared_lib
    make install-shared
  fi
fi

if [ "$1" != "release" ]; then
  if [ -d /mnt/faiss ]; then
    cd /mnt/faiss/faiss-1.9.0
    make -C build install
    cp build/c_api/libfaiss_c.so  /usr/local/lib
  else
    mkdir /mnt/faiss
    cd /mnt/faiss
    wget https://github.com/facebookresearch/faiss/archive/refs/tags/v1.9.0.tar.gz
    tar xzvf v1.9.0.tar.gz
    cd /mnt/faiss/faiss-1.9.0
    cmake -B build -DFAISS_ENABLE_GPU=OFF -DFAISS_ENABLE_PYTHON=OFF -DFAISS_ENABLE_C_API=ON -DBUILD_SHARED_LIBS=ON -DCMAKE_BUILD_TYPE=Release .
    pushd faiss
    ln -s ../perf_tests/ . # Fix bug
    popd
    make -C build install
    cp build/c_api/libfaiss_c.so  /usr/local/lib
  fi
fi

ldconfig

if [ "$1" != "release" ]; then
  if [ -d /mnt/protoc ]; then
    cd /mnt/protoc
    cp bin/protoc /usr/local/bin/
  else
    mkdir /mnt/protoc
    cd /mnt/protoc

    if [ `uname -m` == "arm64" ]; then
      PROTOC_ZIP=protoc-29.3-linux-aarch_64.zip
    fi
    if [ `uname -m` == "x86_64" ]; then
      PROTOC_ZIP=protoc-29.3-linux-x86_64.zip
    fi
    wget https://github.com/protocolbuffers/protobuf/releases/download/v29.3/${PROTOC_ZIP}
    unzip ${PROTOC_ZIP}
    cp bin/protoc /usr/local/bin/
    cp -r include/google /usr/local/include/
  fi
fi

mkdir -p /usr/local/faissdb/bin /usr/local/faissdb/tmp

if [ "$1" == "release" ]; then
  cp -P /mnt/local/lib/* /usr/local/lib/
  ldconfig
  cp -f /mnt/faissdb-build/server/faissdb /usr/local/faissdb/bin/faissdb
  if [ "$2" == "ci" ]; then
    cp -r /mnt/local/include/* /usr/local/include/
    cp /mnt/local/bin/* /usr/local/bin/
    cp /mnt/.faissdb /usr/local/
    eval "$(goenv init -)"
    git clone https://github.com/nodenv/nodenv.git /usr/local/nodenv
    git clone https://github.com/nodenv/node-build.git /usr/local/nodenv/plugins/node-build
    export PATH=/usr/local/nodenv/bin:$PATH
    export NODENV_ROOT=/usr/local/nodenv
    eval "$(nodenv init -)"
    nodenv install 20.18.1
    nodenv global 20.18.1
  fi
else
  eval "$(goenv init -)"
  mkdir -p /mnt/go
  mkdir -p `dirname $GOPATH`
  ln -s /mnt/go $GOPATH
  bash `dirname $0`/make.sh
fi
