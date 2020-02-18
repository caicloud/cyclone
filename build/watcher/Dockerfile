FROM alpine:3.8

LABEL maintainer="zhujian@caicloud.io"

WORKDIR /workspace

RUN apk update && \
    apk add --no-cache ca-certificates bash

COPY ./build/watcher/pvc-watcher.sh /workspace/pvc-watcher.sh

CMD ["/workspace/pvc-watcher.sh"]