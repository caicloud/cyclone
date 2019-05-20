FROM newtmitch/sonar-scanner:3.3.0-alpine

LABEL maintainer="chende@caicloud.io"

ENV WORKDIR /workspace
WORKDIR $WORKDIR

RUN wget -O /bin/jq https://github.com/stedolan/jq/releases/download/jq-1.6/jq-linux64 && \
    chmod +x /bin/jq

COPY ./build/cicd/sonarqube/entrypoint.sh /

ENTRYPOINT ["/entrypoint.sh"]
