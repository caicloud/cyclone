FROM alpine/git:1.0.4

MAINTAINER chende@caicloud.io

ENV WORKDIR /workspace
WORKDIR $WORKDIR

COPY ./build/resolver/git/entrypoint.sh /

ENTRYPOINT ["/entrypoint.sh"]

CMD ["help"]