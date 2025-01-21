#!/usr/bin/env bash
cd `dirname $0`
. definition.sh
docker rmi "${RELEASE_IMAGE}"
set -e

docker image build -t "${RELEASE_IMAGE}" .
