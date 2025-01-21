## Simple primary
### Start faissdb
#### Set ENV
```
export FAISSDB_ROOT=[faissdb repository path]
export FAISSDB_VERSION=0.2.0
```

#### Get docker image
```
docker pull crumbjp/faissdb:$FAISSDB_VERSION
```

#### Start process
```
rm -rf /tmp/faissdb
mkdir /tmp/faissdb
mkdir /tmp/faissdb/log
mkdir /tmp/faissdb/data
mkdir /tmp/faissdb/tmp
docker run \
 --rm \
 -v $FAISSDB_ROOT/config/example:/usr/local/faissdb/conf \
 -v /tmp/faissdb/log:/usr/local/faissdb/log \
 -v /tmp/faissdb/data:/usr/local/faissdb/data \
 -v /tmp/faissdb/tmp:/usr/local/faissdb/tmp \
 -p 9091:9091 \
 -p 20021:20021 \
 -p 21021:21021 \
 crumbjp/faissdb:$FAISSDB_VERSION \
 /usr/local/faissdb/bin/faissdb /usr/local/faissdb/conf/config1.yml
```

#### Replica setting
```
#for Mac
curl -v http://localhost:9091/replicaset -XPUT -d '{"replica": "rs", "members": [{"id": 1, "host": "host.docker.internal:21021", "primary": true}]}'
```

```
#for Linux
curl -v http://localhost:9091/replicaset -XPUT -d '{"replica": "rs", "members": [{"id": 1, "host": "localhost:21021", "primary": true}]}'
```


### Test by faissdb_client
```
git clone https://github.com/crumbjp/faissdb.git
cd faissdb/nodejs
npm install
NODE_PATH=src node ./example/index.js
```
