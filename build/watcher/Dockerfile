FROM alpine:3.8

LABEL maintainer="chende@caicloud.io"

WORKDIR /workspace

COPY ./build/watcher/pvc-watcher.sh /workspace/pvc-watcher.sh

CMD ["/workspace/pvc-watcher.sh"]