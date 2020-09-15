FROM caicloud/cyclone-base-alpine:v1.1.0-arm64v8

LABEL maintainer="zhujian@caicloud.io"

ENV WORKDIR /workspace

WORKDIR $WORKDIR



COPY ./build/resolver/image/entrypoint.sh /

ENTRYPOINT ["/entrypoint.sh"]

CMD ["help"]