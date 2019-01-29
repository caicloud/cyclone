FROM alpine:3.8

LABEL maintainer="zhujian@caicloud.io"

WORKDIR /workspace

RUN apk update && apk add ca-certificates &&\
    apk add curl
#    curl https://storage.googleapis.com/kubernetes-release/release/v1.12.2/bin/linux/amd64/kubectl -o /usr/local/bin/kubectl && \
#    chmod +x /usr/local/bin/kubectl && \
#    kubectl version --client

ENV DOCKER_VERSION=18.06.0
RUN curl -O https://download.docker.com/linux/static/stable/x86_64/docker-${DOCKER_VERSION}-ce.tgz && \
    tar -xzf docker-${DOCKER_VERSION}-ce.tgz && \
    mv docker/docker /usr/local/bin/docker && \
    rm -rf ./docker docker-${DOCKER_VERSION}-ce.tgz


COPY ./bin/workflow/coordinator /workspace/coordinator

CMD ["./coordinator"]