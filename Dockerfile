FROM alpine:edge

RUN apk add --no-cache alpine-sdk tar go

ENV CONSUL_NOTIFY_VERSION 1.0.1
ENV GOPATH /go
ENV GOREPO github.com/allen13/consul-notify
RUN mkdir -p $GOPATH/src/$GOREPO
COPY . $GOPATH/src/$GOREPO
WORKDIR $GOPATH/src/$GOREPO



RUN CGO_ENABLED=0 go build -buildmode=exe consul-notify.go
RUN tar -czvf consul-notify_${CONSUL_NOTIFY_VERSION}_linux_amd64.tar.gz consul-notify
VOLUME /output

CMD cp consul-notify_${CONSUL_NOTIFY_VERSION}_linux_amd64.tar.gz /output
