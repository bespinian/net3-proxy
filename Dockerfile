FROM golang:1 as builder
RUN groupadd -r net3-http-proxy && useradd --no-log-init -r -g net3-http-proxy net3-http-proxy
WORKDIR /usr/src/net3-http-proxy
COPY . ./
RUN make install

FROM scratch
COPY --from=builder /usr/src/net3-http-proxy/bin/net3-http-proxy /usr/net3-http-proxy/
COPY --from=builder /etc/passwd /etc/passwd
USER net3-http-proxy
WORKDIR /usr/net3-http-proxy
ENTRYPOINT ["./net3-http-proxy"]
