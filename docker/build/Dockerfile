FROM ubuntu:24.04
ENV container docker
ENV PATH "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/snap/bin"
ENV LC_ALL C.UTF-8

RUN apt update
RUN apt -y install build-essential cmake git curl vim unzip wget zip libopenblas-dev libsnappy-dev liblz4-dev libzstd-dev zlib1g-dev libbz2-dev libgflags-dev

STOPSIGNAL SIGRTMIN+3
CMD ["/usr/local/faissdb/bin/faissdb", "/usr/local/faissdb/conf/config.yml"]
