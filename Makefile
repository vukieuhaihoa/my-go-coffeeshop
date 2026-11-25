CONFIG_PATH ?= cmd/product/config.yml
PROXY_CONFIG_PATH ?= cmd/proxy/config.yml

.PHONY: run-product build-product run-proxy build-proxy
run-product:
	go run ./cmd/product/ -config $(CONFIG_PATH)

build-product:
	go build -o bin/product ./cmd/product

run-proxy:
	go run ./cmd/proxy/ -config $(PROXY_CONFIG_PATH)

build-proxy:
	go build -o bin/proxy ./cmd/proxy

.PHONY: tidy lint fmt test
tidy:
	go mod tidy

lint:
	golangci-lint run ./...

# chạy gofumpt và goimports theo .golangci.yml
fmt:
	golangci-lint fmt ./...

test:
	go test -race -cover ./...

.PHONY: compose-up compose-down docker-build-product docker-compose-logs
compose-up:
	docker compose up -d

compose-down:
	docker compose down --remove-orphans

docker-build-product:
	docker compose build product

docker-compose-logs:
	docker compose logs -f product

.PHONY: clean
clean:
	rm -rf bin

.PHONY: proto-fmt proto-lint proto-gen proto-breaking
proto-fmt:
	buf format -w

proto-lint:
	buf lint

proto-gen:
	buf generate

proto-breaking:
	buf breaking --against '.git#branch=main'
