FROM  golang:1.6-alpine

EXPOSE 7099

WORKDIR /root

RUN apk add --no-cache ca-certificates tzdata && \
    ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

# Copy cyclone
COPY ./cyclone-server /cyclone-server
COPY ./http/web /http/web
COPY ./notify/provider /template
COPY ./node_modules /root/node_modules

CMD ["/cyclone-server"]
