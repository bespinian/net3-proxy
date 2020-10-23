FROM golang:1-alpine
WORKDIR /usr/src/net3-proxy
COPY . ./
RUN go build -o bin/net3-proxy ./cmd/net3-proxy
ENTRYPOINT ["./bin/net3-proxy"]
