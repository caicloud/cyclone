FROM alpine:3.8

LABEL maintainer="zhujian@caicloud.io"

ENV WORKDIR /workspace
WORKDIR $WORKDIR

RUN apk update && \
    apk add --no-cache ca-certificates bash coreutils subversion

COPY ./build/resolver/svn/entrypoint.sh /

ENTRYPOINT ["/entrypoint.sh"]

CMD ["help"]