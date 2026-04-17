#!/usr/bin/env bash
# Runs the nodejs client mocha suite inside the faissdb-build container.
# Installs nodenv + Node 22.14.0 on demand; goenv is already set up by
# /etc/profile via `bash --login`.
set -eux

rm -rf /tmp/faissdb1 /tmp/faissdb2 /tmp/faissdb3 2>/dev/null || true
mkdir -p /tmp/faissdb1/log /tmp/faissdb1/tmp /tmp/faissdb1/data
mkdir -p /tmp/faissdb2/log /tmp/faissdb2/tmp /tmp/faissdb2/data
mkdir -p /tmp/faissdb3/log /tmp/faissdb3/tmp /tmp/faissdb3/data

if [ ! -d /usr/local/nodenv ]; then
  git clone https://github.com/nodenv/nodenv.git /usr/local/nodenv
  git clone https://github.com/nodenv/node-build.git /usr/local/nodenv/plugins/node-build
fi
export NODENV_ROOT=/usr/local/nodenv
export PATH=$NODENV_ROOT/bin:$PATH
eval "$(nodenv init -)"
if [ ! -d /usr/local/nodenv/versions/22.14.0 ]; then
  nodenv install 22.14.0
fi
nodenv global 22.14.0

cd /mnt/faissdb/nodejs
npm install
bash mocha.sh
