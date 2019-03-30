FROM ubuntu:18.10

LABEL maintainer="chende@caicloud.io"

ENV WORKDIR /workspace
WORKDIR $WORKDIR

RUN apt-get update \
    && apt-get install -y git \
    && rm -rf /var/lib/apt/lists/*

COPY ./build/resolver/git/entrypoint.sh /

ENTRYPOINT ["/entrypoint.sh"]

CMD ["help"]