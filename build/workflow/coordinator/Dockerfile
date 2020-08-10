FROM caicloud/cyclone-base-alpine:v1.1.0

LABEL maintainer="zhujian@caicloud.io"

WORKDIR /workspace

#    curl https://storage.googleapis.com/kubernetes-release/release/v1.12.2/bin/linux/amd64/kubectl -o /usr/local/bin/kubectl && \
#    chmod +x /usr/local/bin/kubectl && \
#    kubectl version --client

COPY ./bin/workflow/coordinator /workspace/coordinator

CMD ["./coordinator"]