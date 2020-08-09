FROM caicloud/cyclone-base-alpine:v1.1.0-arm64v8

LABEL maintainer="zhujian@caicloud.io"

WORKDIR /workspace

COPY ./build/watcher/pvc-watcher.sh /workspace/pvc-watcher.sh

CMD ["/workspace/pvc-watcher.sh"]