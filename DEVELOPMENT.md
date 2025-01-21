## Build faissdb:build
```
git clone https://github.com/crumbjp/faissdb.git
cd faissdb/docker
./build.sh
```

### Start container
```
cd docker
./dev.sh start_container
```

### Start process
```
cd docker
./dev.sh start
```

### Rebuild
Rebuild server from [current source](https://github.com/crumbjp/faissdb/tree/master/docker/dev.sh#L11).

```
cd docker
./dev.sh rebuild
```
