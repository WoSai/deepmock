FROM golang:1.17-bullseye as builder
WORKDIR /go/src/github.com/wosai/deepmock
COPY . /go/src/github.com/wosai/deepmock
ENV GO111MODULE on
ENV GOPROXY https://goproxy.cn,direct
RUN set -e \
    && apt update -y \
    && apt install -y git \
    && REVISION=`git rev-list -1 HEAD` \
    && go mod download \
    && go build -v -ldflags "-X main.version=$REVISION" -o deepmock cmd/main.go

FROM debian:bullseye
WORKDIR /app
COPY --from=builder /go/src/github.com/wosai/deepmock/deepmock .
COPY entrypoint.sh /usr/bin/

EXPOSE 16600
ENTRYPOINT ["entrypoint.sh"]
CMD ["-server-port", ":16600"]
