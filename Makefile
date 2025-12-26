.PHONY: all build clean proto test docker-build docker-up docker-down

GO := go
PROTOC := protoc
PROTO_DIR := api/proto
PROTO_OUT := api/proto

all: proto build

proto:
	$(PROTOC) --go_out=$(PROTO_OUT) --go_opt=paths=source_relative \
		--go-grpc_out=$(PROTO_OUT) --go-grpc_opt=paths=source_relative \
		-I$(PROTO_DIR) \
		$(PROTO_DIR)/*.proto

build: proto
	$(GO) build -o bin/gateway ./cmd/gateway
	$(GO) build -o bin/asr ./cmd/asr
	$(GO) build -o bin/translator ./cmd/translator
	$(GO) build -o bin/tts ./cmd/tts

clean:
	rm -rf bin/
	rm -f $(PROTO_OUT)/*.pb.go

test:
	$(GO) test -v ./...

run-gateway:
	$(GO) run ./cmd/gateway

run-asr:
	$(GO) run ./cmd/asr

run-translator:
	$(GO) run ./cmd/translator

run-tts:
	$(GO) run ./cmd/tts

docker-build:
	docker compose -f deployments/compose.yaml build

docker-up:
	docker compose -f deployments/compose.yaml up -d

docker-down:
	docker compose -f deployments/compose.yaml down

tidy:
	$(GO) mod tidy

deps:
	$(GO) mod download

lint:
	golangci-lint run ./...
