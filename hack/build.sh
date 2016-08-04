#!/bin/sh

set -ex

checkfmt() {
    local files="$(gofmt -l . | grep -v vendor)"
    if [ -n "$files" ]; then
        echo "$files" >&2
        return 1
    fi
}

lint() {
    gometalinter \
        --disable=gotype \
        --vendor \
        --skip=examples \
        --skip=test \
        --deadline=60s \
        --severity=golint:error \
        --errors \
        ./...
}

build() {
    TAGS="static_build netgo"
    test -n "$1" -a -n "$2"
    export GOOS="$1"
    export GOARCH="$2"
    for fn in $(find examples -name main.go); do
        OUT=$(dirname $fn)
        mkdir -p bin/$GOOS/$GOARCH/$(dirname $OUT)
        CGO_ENABLED=0 go build -o bin/$GOOS/$GOARCH/$OUT \
            -a -tags "$TAGS" -installsuffix netgo \
            -ldflags '-extldflags -static' \
            ./$fn
    done
}

case "$1" in
    lint) lint ;;
    checkfmt) checkfmt ;;
    *) build $@ ;;
esac
