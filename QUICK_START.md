
## Simple primary/secondary
### Start faissdb
```
export FAISSDB_ROOT=[faissdb repository path]
docker pull crumbjp/faissdb:0.0.1-alhpa2

mkdir /tmp/faissdb_primary
mkdir /tmp/faissdb_primary/log
mkdir /tmp/faissdb_primary/data
mkdir /tmp/faissdb_primary/tmp
docker run \
 --rm \
 -v $FAISSDB_ROOT/config/test:/usr/local/faissdb/conf \
 -v /tmp/faissdb_primary/log:/usr/local/faissdb/log \
 -v /tmp/faissdb_primary/data:/usr/local/faissdb/data \
 -v /tmp/faissdb_primary/tmp:/usr/local/faissdb/tmp \
 -p 9090:9090 \
 -p 20021:20021 \
 -p 50021:50021 \
 faissdb:0.0.1-alpha2 \
 /usr/local/faissdb/bin/faissdb /usr/local/faissdb/conf/config.yml.primary

mkdir /tmp/faissdb_secondary
mkdir /tmp/faissdb_secondary/log
mkdir /tmp/faissdb_secondary/data
mkdir /tmp/faissdb_secondary/tmp
docker run \
 --rm \
 -v $FAISSDB_ROOT/config/test:/usr/local/faissdb/conf \
 -v /tmp/faissdb_secondary/log:/usr/local/faissdb/log \
 -v /tmp/faissdb_secondary/data:/usr/local/faissdb/data \
 -v /tmp/faissdb_secondary/tmp:/usr/local/faissdb/tmp \
 -p 9091:9091 \
 -p 20022:20022 \
 -p 50022:50022 \
 faissdb:0.0.1-alpha2 \
 /usr/local/faissdb/bin/faissdb /usr/local/faissdb/conf/config.yml.secondary
```

### Test by faissdb_client
```
git clone https://github.com/crumbjp/faissdb_client_node.git
cd faissdb_client_node
npm install
NODE_PATH=src node ./test/index.js

```