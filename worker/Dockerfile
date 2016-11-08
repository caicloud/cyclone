FROM docker:1.10.1-dind

RUN apk update && apk add git && apk add subversion && \
    apk add ca-certificates && \
    apk add tzdata && \
    ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

COPY ./cyclone-worker /cyclone-worker
COPY ./start.sh /start.sh

CMD ["/start.sh"]
