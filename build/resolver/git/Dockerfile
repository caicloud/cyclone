FROM alpine/git:1.0.4

ENV WORKDIR /workspace
WORKDIR $WORKDIR

RUN mkdir -p /workspace/data
COPY ./build/resolver/git/entrypoint.sh /workspace

ENTRYPOINT ["/workspace/entrypoint.sh"]

CMD ["help"]