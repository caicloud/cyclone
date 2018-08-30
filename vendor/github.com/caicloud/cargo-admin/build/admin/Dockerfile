FROM cargo.caicloudprivatetest.com/caicloud/debian:jessie

MAINTAINER Chen De "<chende@caicloud.io>"

RUN mkdir -p /app/etc

COPY ./bin/cargo-admin /app
COPY ./templates/dockerfiles.yml /app

WORKDIR /app

ENTRYPOINT ["./cargo-admin", "-f", "./etc/cargo-admin-config.json"]