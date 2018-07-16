FROM docker:17.03.1-ce-dind

RUN apk add --no-cache git subversion ca-certificates tzdata && \
    ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

COPY bin/worker /cyclone-worker
COPY build/worker/start.sh /start.sh

CMD ["/start.sh"]
