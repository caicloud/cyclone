FROM alpine:3.8

LABEL maintainer="chende@caicloud.io"

ENV WORKDIR /workspace
ENV DOCKER_VERSION 18.03.1-ce
WORKDIR $WORKDIR

RUN apk add --no-cache curl && \
    set -x && \
    curl -L -o /tmp/docker-${DOCKER_VERSION}.tgz https://download.docker.com/linux/static/stable/x86_64/docker-${DOCKER_VERSION}.tgz && \
    tar -xz -C /tmp -f /tmp/docker-${DOCKER_VERSION}.tgz && \
    mv /tmp/docker/docker /usr/bin && \
    rm -rf /tmp/docker && \
    rm /tmp/docker-${DOCKER_VERSION}.tgz


COPY ./build/resolver/image/entrypoint.sh /

ENTRYPOINT ["/entrypoint.sh"]

CMD ["help"]