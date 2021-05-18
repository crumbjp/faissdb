## Simple primary/secondary/secondary
### Start faissdb
#### Set ENV
```
export FAISSDB_ROOT=[faissdb repository path]
export FAISSDB_VERSION=0.0.4
```

#### Get docker image
```
docker pull crumbjp/faissdb:$FAISSDB_VERSION
```

#### Start process1
```
rm -rf /tmp/faissdb1
mkdir /tmp/faissdb1
mkdir /tmp/faissdb1/log
mkdir /tmp/faissdb1/data
mkdir /tmp/faissdb1/tmp
docker run \
 --rm \
 -v $FAISSDB_ROOT/config/example:/usr/local/faissdb/conf \
 -v /tmp/faissdb1/log:/usr/local/faissdb/log \
 -v /tmp/faissdb1/data:/usr/local/faissdb/data \
 -v /tmp/faissdb1/tmp:/usr/local/faissdb/tmp \
 -p 9091:9091 \
 -p 20021:20021 \
 -p 21021:21021 \
 crumbjp/faissdb:$FAISSDB_VERSION \
 /usr/local/faissdb/bin/faissdb /usr/local/faissdb/conf/config1.yml
```

#### Start process2
```
rm -rf /tmp/faissdb2
mkdir /tmp/faissdb2
mkdir /tmp/faissdb2/log
mkdir /tmp/faissdb2/data
mkdir /tmp/faissdb2/tmp
docker run \
 --rm \
 -v $FAISSDB_ROOT/config/example:/usr/local/faissdb/conf \
 -v /tmp/faissdb2/log:/usr/local/faissdb/log \
 -v /tmp/faissdb2/data:/usr/local/faissdb/data \
 -v /tmp/faissdb2/tmp:/usr/local/faissdb/tmp \
 -p 9092:9092 \
 -p 20022:20022 \
 -p 21022:21022 \
 crumbjp/faissdb:$FAISSDB_VERSION \
 /usr/local/faissdb/bin/faissdb /usr/local/faissdb/conf/config2.yml
```

#### Start process3
```
rm -rf /tmp/faissdb3
mkdir /tmp/faissdb3
mkdir /tmp/faissdb3/log
mkdir /tmp/faissdb3/data
mkdir /tmp/faissdb3/tmp
docker run \
 --rm \
 -v $FAISSDB_ROOT/config/example:/usr/local/faissdb/conf \
 -v /tmp/faissdb3/log:/usr/local/faissdb/log \
 -v /tmp/faissdb3/data:/usr/local/faissdb/data \
 -v /tmp/faissdb3/tmp:/usr/local/faissdb/tmp \
 -p 9093:9093 \
 -p 20023:20023 \
 -p 21023:21023 \
 crumbjp/faissdb:$FAISSDB_VERSION \
 /usr/local/faissdb/bin/faissdb /usr/local/faissdb/conf/config3.yml
```

#### Replica setting
```
curl -v http://localhost:9091/replicaset -XPUT -d '{"replica": "rs", "members": [{"id": 1, "host": "host.docker.internal:21021", "primary": true}, {"id": 2, "host": "host.docker.internal:21022", "primary": false}, {"id": 3, "host": "host.docker.internal:21023", "primary": false}]}'
```


### Test by faissdb_client
```
git clone https://github.com/crumbjp/faissdb_client_node.git
cd faissdb_client_node
npm install
NODE_PATH=src node ./test/index.js

```
