FROM alpine:3.6

RUN apk --no-cache add ca-certificates tzdata && \
    ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

COPY ./devops-admin /devops-admin

EXPOSE 7088

ENTRYPOINT ["/bin/sh", "-c", "/devops-admin --cyclone-server ${CYCLONE_SERVER} --mongo-servers ${MONGO_SERVERS}"]
