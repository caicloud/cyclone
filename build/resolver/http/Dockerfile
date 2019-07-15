FROM alpine:3.8

LABEL maintainer="zhujian@caicloud.io"

ENV WORKDIR /workspace
WORKDIR $WORKDIR

RUN apk add --no-cache curl bash

COPY ./build/resolver/http/entrypoint.sh /

ENTRYPOINT ["/entrypoint.sh"]

CMD ["help"]