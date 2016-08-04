FROM golang:1.6-alpine
RUN apk update && apk add curl git tar zip && rm -fr /var/cache/apk/* && \
    go get -v github.com/alecthomas/gometalinter && \
    go get -v golang.org/x/tools/cmd/... && \
    go get -v github.com/FiloSottile/gvt && \
    go get -v github.com/smartystreets/goconvey && \
    go get -v github.com/spf13/hugo && \
    gometalinter --install && \
    chmod -R a+rw /go
