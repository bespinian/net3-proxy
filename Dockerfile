FROM docker.io/library/golang:1 AS builder
WORKDIR /usr/src/app
COPY . ./
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -a -installsuffix cgo -o bin/net3-proxy ./cmd/net3-proxy

FROM docker.io/library/alpine:latest
WORKDIR /usr/src/app
RUN addgroup -S net3proxy && adduser -S net3proxy -G net3proxy
RUN apk add --no-cache \
  ca-certificates \
  dumb-init
COPY --from=builder --chown=net3proxy:net3proxy /usr/src/app/bin/net3-proxy ./
USER net3proxy
EXPOSE 8080
ENTRYPOINT [ "/usr/bin/dumb-init", "--" ]
CMD [ "/usr/src/app/net3-proxy" ]
