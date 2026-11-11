CONFIG_PATH ?= cmd/product/config.yml

.PHONY: run-product build-product
run-product:
	go run ./cmd/product/ -config $(CONFIG_PATH)

build-product:
	go build -o bin/product ./cmd/product

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
