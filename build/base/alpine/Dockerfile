FROM alpine:3.12
LABEL maintainer="zhujian@caicloud.io"

# git version: 2.26.2
RUN apk update && \
    apk add --no-cache ca-certificates bash coreutils git subversion curl jq

ENV DOCKER_VERSION 18.06.0
RUN curl -O https://download.docker.com/linux/static/stable/x86_64/docker-${DOCKER_VERSION}-ce.tgz && \
    tar -xzf docker-${DOCKER_VERSION}-ce.tgz && \
    mv docker/docker /usr/local/bin/docker && \
    rm -rf ./docker docker-${DOCKER_VERSION}-ce.tgz

RUN apk add tzdata && \
    ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone