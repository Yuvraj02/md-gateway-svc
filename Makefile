BIN=bin/gateway
DOCKER_IMAGE?=marketing-digest-gateway
WS_ROOT:=$(abspath .)

.PHONY: build test lint run docker-build tidy

build:
	mkdir -p bin
	go build -o $(BIN) ./cmd/server

test:
	go test ./...

lint:
	go vet ./...

tidy:
	go mod tidy

run: build
	set -a && [ -f .env ] && . ./.env; set +a; ./$(BIN)

docker-build:
	docker build -t $(DOCKER_IMAGE) $(WS_ROOT)
