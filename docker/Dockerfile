FROM ubuntu:24.04
ENV container=docker
ENV PATH="/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/snap/bin"
ENV LC_ALL=C.UTF-8

RUN apt update
RUN apt install -y libopenblas0 libsnappy1v5 liblz4-1 libzstd1 zlib1g libbz2-1.0 libgomp1 libgflags2.2 curl

RUN --mount=type=bind,source=build/mnt,target=/mnt,rw \
    --mount=type=bind,from=faissdb,source=.,target=/mnt/faissdb,ro \
    --mount=type=bind,from=gofaiss,source=.,target=/mnt/go-faiss,ro \
    bash --login /mnt/build.sh release

STOPSIGNAL SIGRTMIN+3
CMD ["/usr/local/faissdb/bin/faissdb", "/usr/local/faissdb/conf/config.yml"]
