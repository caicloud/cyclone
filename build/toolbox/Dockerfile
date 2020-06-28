FROM caicloud/cyclone-base-alpine:v1.0.0

LABEL maintainer="zhujian@caicloud.io"

WORKDIR /usr/bin/cyclone-toolbox

RUN apk add --no-cache coreutils

COPY ./build/toolbox/resolver-runner.sh ./bin/toolbox/fstream /usr/bin/cyclone-toolbox/
