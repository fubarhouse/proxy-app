FROM golang:latest AS builder

LABEL stage=builder
RUN mkdir -p /go/src/github.com/skpr/proxy-app
COPY . /go/src/github.com/skpr/proxy-app
WORKDIR /go/src/github.com/skpr/proxy-app
RUN GO111MODULE=on go mod vendor
RUN GO111MODULE=on go mod verify
RUN GO111MODULE=on GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static"' -o proxy-app .

FROM skpr/base:1.x

ENV PROXY_APP_ADDR=":8080"

COPY --from=builder /go/src/github.com/skpr/proxy-app/proxy-app /usr/local/bin/proxy-app
RUN chmod +x /usr/local/bin/proxy-app

CMD ["/usr/local/bin/proxy-app"]
