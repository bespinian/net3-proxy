FROM golang:1-alpine
WORKDIR /usr/src/net3-http-proxy
COPY . ./
RUN go build -o bin/net3-http-proxy ./cmd/net3-http-proxy
ENTRYPOINT ["./bin/net3-http-proxy"]
