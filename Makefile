.PHONY: build
all:
	go build -o bin/net3-http-proxy ./cmd/net3-http-proxy

.PHONY: build-image
all:
	docker build -t bespinian/net3-http-proxy .

.PHONY: lint
lint:
	golangci-lint run

.PHONY: test
test:
	go test -race -cover ./...

.PHONY: clean
clean:
	rm -rf bin
