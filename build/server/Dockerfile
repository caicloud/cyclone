FROM alpine:3.8

LABEL maintainer="zhujian@caicloud.io"

WORKDIR /root

RUN apk update && apk add ca-certificates && \
    apk add --no-cache subversion && \
    apk add tzdata && \
    ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

# Copy cyclone server and stage templates
COPY bin/server /cyclone-server
COPY manifests/templates /root/templates

EXPOSE 7099

ENTRYPOINT ["/cyclone-server"]
