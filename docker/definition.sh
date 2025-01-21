ARCH=`uname -m`
VERSION=`cat version`
MANIFEST="crumbjp/faissdb:${VERSION}"
RELEASE_IMAGE="${MANIFEST}-${ARCH}"
