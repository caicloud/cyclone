FROM golang:cross

# To get zip and apt-transport-https
RUN apt-get update

# buildkite-agent for artifact management
RUN apt-get install -y apt-transport-https
RUN echo deb https://apt.buildkite.com/buildkite-agent unstable main > /etc/apt/sources.list.d/buildkite-agent.list
RUN apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys 32A37959C2FA5C3C99EFBC32A79206696452D198
RUN apt-get update

# Install buildkite-agent
RUN apt-get install -y buildkite-agent

# For dealing with Go deps
RUN go get github.com/tools/godep

# For creating Github releases
RUN go get github.com/buildkite/github-release

# Zip for win and osx releases
RUN apt-get install -y zip

# For creating deb and rpm packages
RUN apt-get install -y ruby ruby-dev rpm
RUN gem install fpm package_cloud

# Install UTF-8 locale for package_cloud
RUN apt-get install -y locales
RUN echo "en_US UTF-8" > /etc/locale.gen && locale-gen en_US.UTF-8
ENV LANG=en_US.UTF-8
ENV LANGUAGE=en_US:en
ENV LC_ALL=en_US.UTF-8

RUN mkdir -p /go/src/github.com/buildkite/terminal
ADD . /go/src/github.com/buildkite/terminal

WORKDIR /go/src/github.com/buildkite/terminal
