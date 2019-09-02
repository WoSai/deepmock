FROM golang:1.12-alpine as builder
WORKDIR /go/src/github.com/qastub/deepmock
COPY . /go/src/github.com/qastub/deepmock
ENV GO111MODULE on
ENV GOPROXY https://proxy.qastub.com
RUN set -e \
    && sed -i "s/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g" /etc/apk/repositories \
    && apk update \
    && apk add git \
    && REVISION=`git rev-list -1 HEAD` \
    && go mod download \
    && go build -v -ldflags "-X main.version=$REVISION" -o deepmock cmd/main.go

FROM alpine
WORKDIR /app
COPY --from=builder /go/src/github.com/qastub/deepmock/deepmock .
COPY entrypoint.sh /usr/bin/
VOLUME /app/log
ENV DEEPMOCK_LOGFILE /app/log/deepmock.log
EXPOSE 16600
ENTRYPOINT ["entrypoint.sh"]
CMD ["-port", ":16600"]