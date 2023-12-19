FROM ubuntu:22.04
ENV container docker
ENV PATH "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/snap/bin"
ENV LANG ja_JP.UTF-8
ENV LC_ALL C.UTF-8

RUN apt update
RUN apt -y install build-essential cmake git curl vim unzip wget zip libopenblas-base libopenblas-dev libsnappy-dev liblz4-dev libzstd-dev libz-dev libbz2-dev

VOLUME ["/sys/fs/cgroup"]
STOPSIGNAL SIGRTMIN+3
CMD ["/sbin/init"]
