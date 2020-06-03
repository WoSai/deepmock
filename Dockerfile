FROM golang:1.14-alpine as builder
WORKDIR /go/src/github.com/wosai/deepmock
COPY . /go/src/github.com/wosai/deepmock
ENV GO111MODULE on
ENV GOPROXY https://goproxy.cn,direct
RUN set -e \
    && sed -i "s/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g" /etc/apk/repositories \
    && apk update \
    && apk add git \
    && REVISION=`git rev-list -1 HEAD` \
    && go mod download \
    && go build -v -ldflags "-X main.version=$REVISION" -o deepmock cmd/main.go

FROM alpine
WORKDIR /app
COPY --from=builder /go/src/github.com/wosai/deepmock/deepmock .
COPY entrypoint.sh /usr/bin/
EXPOSE 16600
ENTRYPOINT ["entrypoint.sh"]
CMD ["-server-port", ":16600"]
