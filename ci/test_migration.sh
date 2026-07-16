#!/usr/bin/env bash
# Verifies live migration between faissdb versions with the released images.
# A 2-node cluster is built on OLD_IMAGE, then each node is switched to
# NEW_IMAGE and back, reusing the same data volumes (in-place upgrade with
# gap-sync replay). The suite covers both mixed-version directions:
# old-primary/new-secondary (old-format oplog consumed by the new binary) and
# new-primary/old-secondary (delta-carrying oplog consumed by the old binary),
# plus rollback and the SetCollections UNIMPLEMENTED behavior on old servers.
#
# Old/new container pairs share named volumes; the mocha suite switches
# versions by starting/stopping containers through the docker API, so the
# driver container gets the host docker socket. All containers join the
# driver's network namespace so localhost endpoints work unchanged.
#
# Usage: ci/test_migration.sh
#   OLD_IMAGE / NEW_IMAGE / CI_IMAGE env vars override the image tags
#   (default: crumbjp/faissdb:0.3.1-<arch> / <docker/version>-<arch> / -ci).
set -eux

cd "$(dirname "$0")/.."
REPO=$(pwd)
VERSION=$(cat docker/version)
ARCH=$(uname -m)
OLD_IMAGE="${OLD_IMAGE:-crumbjp/faissdb:0.3.1-${ARCH}}"
NEW_IMAGE="${NEW_IMAGE:-crumbjp/faissdb:${VERSION}-${ARCH}}"
CI_IMAGE="${CI_IMAGE:-crumbjp/faissdb:${VERSION}-ci}"
DRIVER='faissdb-mig-driver'
NODES='faissdb-mig-node1-old faissdb-mig-node1-new faissdb-mig-node2-old faissdb-mig-node2-new'
VOLUMES='faissdb-mig1-log faissdb-mig1-tmp faissdb-mig1-data faissdb-mig2-log faissdb-mig2-tmp faissdb-mig2-data'

cleanup() {
  docker rm -f ${DRIVER} ${NODES} 2>/dev/null || true
  docker volume rm ${VOLUMES} 2>/dev/null || true
}
cleanup
if [ -z "${KEEP:-}" ]; then
  trap cleanup EXIT
fi

docker run -d --name "${DRIVER}" --tmpfs /tmp:exec \
  -v "${REPO}:/repo:ro" \
  -v /var/run/docker.sock:/var/run/docker.sock \
  "${CI_IMAGE}" sleep infinity

for n in 1 2; do
  for ver in old new; do
    if [ "$ver" == "old" ]; then
      image="${OLD_IMAGE}"
    else
      image="${NEW_IMAGE}"
    fi
    docker create --name "faissdb-mig-node${n}-${ver}" --network "container:${DRIVER}" \
      -v "${REPO}/config/test:/conf:ro" \
      -v "faissdb-mig${n}-log:/tmp/faissdb${n}/log" \
      -v "faissdb-mig${n}-tmp:/tmp/faissdb${n}/tmp" \
      -v "faissdb-mig${n}-data:/tmp/faissdb${n}/data" \
      "${image}" \
      "/usr/local/faissdb/bin/faissdb" "/conf/config${n}.yml"
  done
done

docker exec "${DRIVER}" bash --login -c '
set -eux
. /usr/local/.faissdb
mkdir -p /ws
cp -a /repo/nodejs /ws/nodejs
cp -a /repo/config /ws/config
cd /ws/nodejs
npm install
bash mocha.sh --timeout 240000 test/migration.js
'
