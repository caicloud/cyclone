FROM docker:18.03.1-ce-dind

# Set sonar-scanner version.
ENV SONAR_SCANNER_VERSION 3.2.0.1227

RUN apk add --no-cache git subversion ca-certificates tzdata curl openjdk8-jre && \
    ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

RUN curl --insecure -o ./sonarscanner.zip -L https://binaries.sonarsource.com/Distribution/sonar-scanner-cli/sonar-scanner-cli-${SONAR_SCANNER_VERSION}-linux.zip && \
    unzip sonarscanner.zip && \
    rm sonarscanner.zip && \
    mv sonar-scanner-${SONAR_SCANNER_VERSION}-linux sonar-scanner && \
    sed -i 's/use_embedded_jre=true/use_embedded_jre=false/g' /sonar-scanner/bin/sonar-scanner

ENV SONAR_RUNNER_HOME /sonar-scanner
ENV PATH ${PATH}:/sonar-scanner/bin

COPY bin/worker /cyclone-worker
COPY build/worker/start.sh /start.sh

CMD ["/start.sh"]
