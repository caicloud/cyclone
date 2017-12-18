FROM alpine:3.6

WORKDIR /root

RUN apk update && apk add ca-certificates && \
    apk add tzdata && \
    ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

# Copy cyclone
COPY bin/server /cyclone-server
COPY ./http/web /http/web
COPY ./notify/provider /template
COPY ./node_modules /root/node_modules
COPY ./api/templates /templates

EXPOSE 7099
EXPOSE 8000

ENTRYPOINT ["/cyclone-server"]
