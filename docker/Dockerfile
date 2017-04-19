FROM busybox:ubuntu-14.04

MAINTAINER tsaikd <tsaikd@gmail.com>

ENV GOLASTICDUMP_VERSION 0.0.2

ADD https://github.com/tsaikd/golasticdump/releases/download/${GOLASTICDUMP_VERSION}/golasticdump-Linux-x86_64 /usr/local/bin/golasticdump

RUN chmod +x /usr/local/bin/golasticdump

ENTRYPOINT ["golasticdump"]
