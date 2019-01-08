FROM alpine:3.8

ENV WORKDIR /workspace
WORKDIR $WORKDIR

RUN mkdir -p /workspace/data && \
    apk add --no-cache curl

COPY ./build/resolver/kv/entrypoint.sh /workspace

ENTRYPOINT ["/workspace/entrypoint.sh"]

CMD ["help"]