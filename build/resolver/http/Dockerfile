FROM caicloud/cyclone-base-alpine:v1.1.0

LABEL maintainer="zhujian@caicloud.io"

ENV WORKDIR /workspace
WORKDIR $WORKDIR

COPY ./build/resolver/http/entrypoint.sh /

ENTRYPOINT ["/entrypoint.sh"]

CMD ["help"]