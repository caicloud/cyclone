FROM ubuntu:18.10

LABEL maintainer="chende@caicloud.io"

ENV WORKDIR /workspace
WORKDIR $WORKDIR

RUN apt-get -y update \
    && apt-get install -y subversion locales \
    && locale-gen en_US.UTF-8 \
    && apt-get -s clean \
    && rm -rf /var/lib/apt/lists/*

ENV LANG='en_US.UTF-8' LANGUAGE='en_US.UTF-8' LC_ALL='en_US.UTF-8'

COPY ./build/resolver/svn/entrypoint.sh /

ENTRYPOINT ["/entrypoint.sh"]

CMD ["help"]