#!/usr/bin/env bash
cd `dirname $0`
. definition.sh
CONTAINER_NAME='faissdb-dev'

function start_container {
  mkdir -p build/mnt/data
  mkdir -p build/mnt/log
  docker run --name="${CONTAINER_NAME}" -ti --tmpfs /run --tmpfs /run/lock --tmpfs /tmp:exec \
   -v `pwd`/build/mnt:/mnt \
   -v `pwd`/..:/mnt/faissdb \
   -v `pwd`/build/mnt/data:/usr/local/faissdb/data \
   -v `pwd`/build/mnt/log:/usr/local/faissdb/log \
   -v `pwd`/../nodejs/example:/usr/local/faissdb/conf \
   -p 9091:9091 \
   -p 20021:20021 \
   -p 21021:21021 \
   -d $1 \
   /bin/bash
}

if [ "$1" == "start_container" ]; then
  start_container faissdb:build
fi

if [ "$1" == "stop_container" ]; then
  docker rm -f "${CONTAINER_NAME}"
fi

if [ "$1" == "rebuild" ]; then
  docker exec "${CONTAINER_NAME}" bash --login /mnt/make.sh
fi

if [ "$1" == "start" ]; then
  docker exec "${CONTAINER_NAME}" /usr/local/faissdb/bin/faissdb /usr/local/faissdb/conf/config.yml
fi

if [ "$1" == "setup" ]; then
  if [ `uname` == "Linux" ]; then
    curl -v http://localhost:9091/replicaset -XPUT -d '{"replica": "rs", "members": [{"id": 1, "host": "localhost:21021", "primary": true}]}'
  else
    curl -v http://localhost:9091/replicaset -XPUT -d '{"replica": "rs", "members": [{"id": 1, "host": "host.docker.internal:21021", "primary": true}]}'
  fi
fi

if [ "$1" == "stop" ]; then
  docker exec "${CONTAINER_NAME}" kill `docker exec "${CONTAINER_NAME}" cat /usr/local/faissdb/tmp/faissdb.pid`
fi

if [ "$1" == "build_ci_container" ]; then
  docker image build -t "${MANIFEST}-ci" . -f Dockerfile.ci
fi

if [ "$1" == "start_release_container" ]; then
  start_container "${RELEASE_IMAGE}"
fi

if [ "$1" == "start_manifest_container" ]; then
  start_container "${MANIFEST}"
fi

if [ "$1" == "push" ]; then
  docker push "${RELEASE_IMAGE}"
fi

if [ "$1" == "manifest" ]; then
  set -e
  docker manifest create "${MANIFEST}" "${MANIFEST}-x86_64" "${MANIFEST}-arm64" --amend
#  docker manifest annotate --arch amd64 "${MANIFEST}" "${MANIFEST}-x86_64"
#  docker manifest annotate --arch arm64 "${MANIFEST}" "${MANIFEST}-arm64"
  docker manifest inspect "${MANIFEST}"
  docker manifest push ${MANIFEST}
fi
