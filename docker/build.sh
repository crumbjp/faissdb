#!/usr/bin/env bash
BUIILD_CONTAINER='faissdb-build'
BASE_IMAGE='faissdb:base'
BUILD_IMAGE='faissdb:build'
docker rm -f "${BUIILD_CONTAINER}"
docker rmi "${BASE_IMAGE}"
docker rmi "${BUILD_IMAGE}"

set -e
cd `dirname $0`

rm -rf build/mnt
cp -r mnt build/mnt
# cp mnt/* build/mnt/
# cp mnt/.faissdb build/mnt/.faissdb

cd build
docker image build -t "${BASE_IMAGE}" .
docker run --name="${BUIILD_CONTAINER}" -ti --tmpfs /run --tmpfs /run/lock --tmpfs /tmp:exec \
 -v /lib/modules:/lib/modules:ro \
 -v `pwd`/mnt:/mnt \
 -v `pwd`/../..:/mnt/faissdb \
 -d "${BASE_IMAGE}" \
 /bin/bash

docker exec "${BUIILD_CONTAINER}" bash --login /mnt/build.sh

docker commit "${BUIILD_CONTAINER}" "${BUILD_IMAGE}"
if [ "$1" != "keep" ]; then
  docker rm -f "${BUIILD_CONTAINER}"
fi
