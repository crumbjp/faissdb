#!/usr/bin/env bash
set -e

log() {
  echo "======== [$(date '+%H:%M:%S')] $* ========"
}

if [ "$1" == "release" ]; then
  echo Do nothing
else
  log "Setting up /usr/local symlinks"
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
fi

cd /mnt

if [ "$1" != "release" ] || [ "$2" == "ci" ]; then
  GO_VERSION=$(cat /mnt/faissdb/server/.go-version 2>/dev/null || cat /mnt/.go-version)
  echo "${GO_VERSION}" > /mnt/.go-version
  log "Installing Go ${GO_VERSION}"

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

  goenv install ${GO_VERSION}
  goenv global ${GO_VERSION}
  log "Go ${GO_VERSION} installed"
fi

if [ "$1" != "release" ]; then
  log "Building RocksDB"
  if [ -d /mnt/rocksdb ]; then
    cd /mnt/rocksdb
    export PORTABLE=1
    make install-shared
  else
    git clone https://github.com/facebook/rocksdb.git /mnt/rocksdb
    cd /mnt/rocksdb
    git checkout -b v10.9.1 refs/tags/v10.9.1
    export PORTABLE=1
    make shared_lib
    make install-shared
  fi
  log "RocksDB done"
fi

if [ "$1" != "release" ]; then
  log "Building FAISS"
  if [ -d /mnt/faiss ]; then
    cd /mnt/faiss/faiss-1.14.1
    make -C build install
    cp build/c_api/libfaiss_c.so  /usr/local/lib
  else
    mkdir /mnt/faiss
    cd /mnt/faiss
    wget https://github.com/facebookresearch/faiss/archive/refs/tags/v1.14.1.tar.gz
    tar xzvf v1.14.1.tar.gz
    cd /mnt/faiss/faiss-1.14.1
    cmake -B build -DFAISS_ENABLE_GPU=OFF -DFAISS_ENABLE_PYTHON=OFF -DFAISS_ENABLE_C_API=ON -DBUILD_SHARED_LIBS=ON -DCMAKE_BUILD_TYPE=Release .
    make -C build install
    cp build/c_api/libfaiss_c.so  /usr/local/lib
  fi
  log "FAISS done"
fi

ldconfig

if [ "$1" != "release" ]; then
  log "Installing protoc"
  if [ -d /mnt/protoc ]; then
    cd /mnt/protoc
    cp bin/protoc /usr/local/bin/
  else
    mkdir /mnt/protoc
    cd /mnt/protoc

    if [ `uname -m` == "arm64" ]; then
      PROTOC_ZIP=protoc-29.3-linux-aarch_64.zip
    fi
    if [ `uname -m` == "aarch64" ]; then
      PROTOC_ZIP=protoc-29.3-linux-aarch_64.zip
    fi
    if [ `uname -m` == "x86_64" ]; then
      PROTOC_ZIP=protoc-29.3-linux-x86_64.zip
    fi
    log "Downloading ${PROTOC_ZIP}"
    wget https://github.com/protocolbuffers/protobuf/releases/download/v29.3/${PROTOC_ZIP}
    unzip ${PROTOC_ZIP}
    cp bin/protoc /usr/local/bin/
    cp -r include/google /usr/local/include/
  fi
  log "protoc done"
fi

mkdir -p /usr/local/faissdb/bin /usr/local/faissdb/tmp

if [ "$1" == "release" ]; then
  log "Release: copying libraries"
  find /mnt/local/lib -mindepth 1 -maxdepth 1 ! -type d ! -name 'libbenchmark*' -exec cp -P {} /usr/local/lib/ \;
  if command -v strip >/dev/null; then
    log "Release: stripping libraries"
    find /usr/local/lib -maxdepth 1 -type f -name '*.so*' -exec strip --strip-unneeded {} \;
  fi
  ldconfig
  cp -f /mnt/faissdb-build/server/faissdb /usr/local/faissdb/bin/faissdb
  if command -v strip >/dev/null; then
    strip /usr/local/faissdb/bin/faissdb
  fi
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
    nodenv install 22.14.0
    nodenv global 22.14.0
    log "CI setup done"
  fi
  log "Release done"
else
  log "Building faissdb"
  eval "$(goenv init -)"
  mkdir -p /mnt/go
  mkdir -p `dirname $GOPATH`
  ln -s /mnt/go $GOPATH
  bash `dirname $0`/make.sh
  log "Build complete"
fi
