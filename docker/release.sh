#!/usr/bin/env bash
cd `dirname $0`
. definition.sh
docker rmi "${RELEASE_IMAGE}"
set -e

docker image build \
  --build-context faissdb=.. \
  --build-context gofaiss=../../go-faiss \
  -t "${RELEASE_IMAGE}" .
