FROM golang:1.7.5-alpine

# change path to your repo dir
WORKDIR /go/src/path/to/your/app
ENV GOPATH /go

# change path to your repo dir
COPY . /go/src/path/to/your/app

RUN go get -d -v &&  go test -v . && go install -v

# change app to your app name
CMD ["app"]
