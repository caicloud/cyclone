FROM caicloud/cyclone-base-alpine:v1.0.0

LABEL maintainer="zhujian@caicloud.io"

ENV WORKDIR /workspace
WORKDIR $WORKDIR

COPY ./build/resolver/svn/entrypoint.sh /

ENTRYPOINT ["/entrypoint.sh"]

CMD ["help"]