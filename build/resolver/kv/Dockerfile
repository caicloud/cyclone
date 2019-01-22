FROM alpine:3.8

MAINTAINER chende@caicloud.io

ENV WORKDIR /workspace
WORKDIR $WORKDIR

RUN mkdir -p /workspace/data && \
    apk add --no-cache curl

COPY ./build/resolver/kv/entrypoint.sh /

ENTRYPOINT ["/entrypoint.sh"]

CMD ["help"]