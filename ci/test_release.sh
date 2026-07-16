#!/usr/bin/env bash
# E2E-tests a release image as shipped: the release containers only run the
# bundled binary (no test tooling injected), while the mocha suite runs in a
# driver container built from the CI image and connects over the network.
# The release containers join the driver's network namespace so the suite's
# localhost endpoints work unchanged. The suite starts cluster nodes on
# demand ('Build cluster' / 'Start new node'), so the driver gets the host
# docker socket and FAISSDB is overridden with a wrapper that starts the
# corresponding pre-created release container via the docker API.
#
# Usage: ci/test_release.sh
#   RELEASE_IMAGE / CI_IMAGE env vars override the image tags
#   (default: crumbjp/faissdb:<docker/version>-<arch> / -ci).
set -eux

cd "$(dirname "$0")/.."
REPO=$(pwd)
VERSION=$(cat docker/version)
ARCH=$(uname -m)
RELEASE_IMAGE="${RELEASE_IMAGE:-crumbjp/faissdb:${VERSION}-${ARCH}}"
CI_IMAGE="${CI_IMAGE:-crumbjp/faissdb:${VERSION}-ci}"
DRIVER='faissdb-release-driver'
NODE_PREFIX='faissdb-release-test-'

cleanup() {
  docker rm -f "${DRIVER}" "${NODE_PREFIX}1" "${NODE_PREFIX}2" "${NODE_PREFIX}3" 2>/dev/null || true
}
cleanup
trap cleanup EXIT

docker run -d --name "${DRIVER}" --tmpfs /tmp:exec \
  -v "${REPO}:/repo:ro" \
  -v /var/run/docker.sock:/var/run/docker.sock \
  "${CI_IMAGE}" sleep infinity

for n in 1 2 3; do
  docker create --name "${NODE_PREFIX}${n}" --network "container:${DRIVER}" \
    -v "${REPO}/config/test:/conf:ro" \
    -v "/tmp/faissdb${n}/log" -v "/tmp/faissdb${n}/tmp" -v "/tmp/faissdb${n}/data" \
    "${RELEASE_IMAGE}" \
    "/usr/local/faissdb/bin/faissdb" "/conf/config${n}.yml"
done

docker exec "${DRIVER}" bash -c 'cat > /usr/local/bin/faissdb-release-node <<"EOF"
#!/usr/bin/env bash
n=$(basename "$1" | tr -dc "0-9")
curl -s --unix-socket /var/run/docker.sock -X POST "http://localhost/containers/faissdb-release-test-${n}/start"
EOF
chmod +x /usr/local/bin/faissdb-release-node'

docker exec "${DRIVER}" bash --login -c '
set -eux
. /usr/local/.faissdb
mkdir -p /ws
cp -a /repo/nodejs /ws/nodejs
cp -a /repo/config /ws/config
cd /ws/nodejs
npm install
FAISSDB=/usr/local/bin/faissdb-release-node bash mocha.sh --timeout 240000 test/index.js
'
