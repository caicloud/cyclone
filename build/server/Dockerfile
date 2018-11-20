FROM alpine:3.6

WORKDIR /root

RUN apk update && apk add ca-certificates && \
    apk add --no-cache subversion && \
    apk add tzdata && \
    ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

# Copy cyclone
COPY bin/server /cyclone-server

EXPOSE 7099

ENTRYPOINT ["/cyclone-server"]
