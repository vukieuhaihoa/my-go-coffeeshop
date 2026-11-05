CONFIG_PATH ?= cmd/product/config.yml

.PHONY: run-product build-product
run-product:
	go run ./cmd/product/ -config $(CONFIG_PATH)

build-product:
	go build -o bin/product ./cmd/product

.PHONY: tidy lint fmt
tidy:
	go mod tidy

lint:
	golangci-lint run ./...

fmt:
	go fmt ./...

.PHONY: compose-up compose-down
compose-up:
	docker compose up -d

compose-down:
	docker compose down --remove-orphans

.PHONY: clean
clean:
	rm -rf bin
