FROM library/openjdk:8-alpine3.8

RUN apk add --no-cache curl grep sed unzip bash nodejs nodejs-npm

# Set timezone to CST
ENV TZ=America/Chicago
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

WORKDIR /usr/src

RUN curl --insecure -o ./sonarscanner.zip -L https://binaries.sonarsource.com/Distribution/sonar-scanner-cli/sonar-scanner-cli-3.3.0.1492-linux.zip && \
	unzip sonarscanner.zip && \
	rm sonarscanner.zip && \
	mv sonar-scanner-3.3.0.1492-linux /usr/lib/sonar-scanner && \
	ln -s /usr/lib/sonar-scanner/bin/sonar-scanner /usr/local/bin/sonar-scanner

ENV SONAR_RUNNER_HOME=/usr/lib/sonar-scanner

COPY build/base/openjdk/sonar-runner.properties /usr/lib/sonar-scanner/conf/sonar-scanner.properties

#   ensure Sonar uses the provided Java for musl instead of a borked glibc one
RUN sed -i 's/use_embedded_jre=true/use_embedded_jre=false/g' /usr/lib/sonar-scanner/bin/sonar-scanner

# Separating ENTRYPOINT and CMD operations allows for core execution variables to
# be easily overridden by passing them in as part of the `docker run` command.
# This allows the default /usr/src base dir to be overridden by users as-needed.

# ENTRYPOINT ["sonar-scanner"] 
CMD ["-Dsonar.projectBaseDir=/usr/src"]

# ==================== #
# Contents above copied from https://github.com/newtmitch/docker-sonar-scanner/blob/master/Dockerfile.sonarscanner-3.3.0-alpine
# ==================== #

# FROM newtmitch/sonar-scanner:3.3.0-alpine

LABEL maintainer="zhujian@caicloud.io"

RUN apk update && \
    apk add --no-cache ca-certificates bash coreutils curl jq