#!/usr/bin/env bash
BUIILD_CONTAINER='faissdb-build'
BASE_IMAGE='faissdb:base'
BUILD_IMAGE='faissdb:build'

RETRY=false
KEEP=false
for arg in "$@"; do
  case "$arg" in
    retry) RETRY=true ;;
    keep)  KEEP=true ;;
  esac
done

set -e
cd `dirname $0`

if [ "$RETRY" != "true" ]; then
  docker rm -f "${BUIILD_CONTAINER}"
  docker rmi -f "${BASE_IMAGE}"
  docker rmi -f "${BUILD_IMAGE}"
  rm -rf build/mnt
  cp -r mnt build/mnt

  cd build
  docker image build -t "${BASE_IMAGE}" .
  docker run --name="${BUIILD_CONTAINER}" -ti --tmpfs /run --tmpfs /run/lock --tmpfs /tmp:exec \
   -v /lib/modules:/lib/modules:ro \
   -v `pwd`/mnt:/mnt \
   -v `pwd`/../..:/mnt/faissdb \
   -v `pwd`/../../../go-faiss:/mnt/go-faiss \
   -d "${BASE_IMAGE}" \
   /bin/bash
else
  cp mnt/build.sh build/mnt/build.sh
  cd build
fi

docker exec "${BUIILD_CONTAINER}" bash --login /mnt/build.sh

docker commit "${BUIILD_CONTAINER}" "${BUILD_IMAGE}"
if [ "$KEEP" != "true" ]; then
  docker rm -f "${BUIILD_CONTAINER}"
fi
