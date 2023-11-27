FROM golang:1.21-bullseye as builder
WORKDIR /go/src/
COPY . /go/src/
ENV GOPROXY https://goproxy.cn,direct
RUN set -e \
    && apt update -y \
    && apt install -y git \
    && REVISION=`git rev-list -1 HEAD` \
    && go mod download \
    && go build -v -ldflags "-X main.version=$REVISION" -o deepmock cmd/main.go

FROM debian:bullseye
WORKDIR /app
COPY --from=builder /go/src/deepmock .
COPY entrypoint.sh /usr/bin/

EXPOSE 16600
ENTRYPOINT ["entrypoint.sh"]
CMD ["-server-port", ":16600"]
