FROM ubuntu:24.04 AS staging
ENV LC_ALL=C.UTF-8

RUN --mount=type=bind,source=build/mnt,target=/mnt,rw \
    --mount=type=bind,from=faissdb,source=.,target=/mnt/faissdb,ro \
    --mount=type=bind,from=gofaiss,source=.,target=/mnt/go-faiss,ro \
    apt update \
 && apt install -y --no-install-recommends libopenblas0 libsnappy1v5 liblz4-1 libzstd1 zlib1g libbz2-1.0 libgomp1 libgflags2.2 binutils \
 && bash --login /mnt/build.sh release rootfs

FROM scratch
COPY --from=staging /out /
STOPSIGNAL SIGRTMIN+3
CMD ["/usr/local/faissdb/bin/faissdb", "/usr/local/faissdb/conf/config.yml"]
