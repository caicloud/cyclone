FROM caicloud/cyclone-base-alpine:v1.1.0

LABEL maintainer="zhujian@caicloud.io"

WORKDIR /root

# Copy cyclone server and stage templates
COPY bin/server /cyclone-server
COPY manifests/templates /root/templates

EXPOSE 7099

ENTRYPOINT ["/cyclone-server"]
