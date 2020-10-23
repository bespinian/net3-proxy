.PHONY: build
all:
	go build -o bin/net3-proxy ./cmd/net3-proxy

.PHONY: build-image
build-image:
	docker build -t bespinian/net3-proxy .

.PHONY: lint
lint:
	golangci-lint run

.PHONY: test
test:
	go test -race -cover ./...

.PHONY: clean
clean:
	rm -rf bin
