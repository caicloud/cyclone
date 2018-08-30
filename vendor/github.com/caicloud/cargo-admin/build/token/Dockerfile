FROM cargo.caicloudprivatetest.com/caicloud/debian:jessie

MAINTAINER Chen De "<chende@caicloud.io>"

RUN mkdir -p /app/etc

COPY ./bin/cargo-token /app

WORKDIR /app

ENTRYPOINT ["./cargo-token", "-f", "./etc/cargo-token-config.json"]