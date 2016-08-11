FROM golang:alpine

ENV CONSUL_NOTIFY_VERSION 0.0.4
ENV GOPATH /go
ENV GOREPO github.com/allen13/consul-notify
RUN mkdir -p $GOPATH/src/$GOREPO
COPY . $GOPATH/src/$GOREPO
WORKDIR $GOPATH/src/$GOREPO

RUN set -ex \
	&& apk add --no-cache git tar\
	&& go get github.com/tools/godep

RUN $GOPATH/bin/godep go build consul-notify.go
RUN tar -czvf consul-notify_${CONSUL_NOTIFY_VERSION}_linux_amd64.tar.gz consul-notify
VOLUME /output

CMD cp consul-notify_${CONSUL_NOTIFY_VERSION}_linux_amd64.tar.gz /output
