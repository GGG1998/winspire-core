BINARY := competition-host-stream
SERVICE_DIR := $(shell pwd)

.PHONY: build run test lint sqlc migrate clean

build:
	go build -o bin/$(BINARY) ./cmd/competition-host-stream

run:
	POSTGRES_DSN?=postgres://postgres:postgres@localhost:5432/competition_host?sslmode=disable
	REDIS_ADDR?=localhost:6379
	go run ./cmd/competition-host-stream

test:
	go test ./...

lint:
	gofmt -w $$(find . -type f -name '*.go')

sqlc:
	sqlc generate

migrate:
	atlas migrate apply \
		--dir file://$(SERVICE_DIR)/migrations \
		--url "$${POSTGRES_DSN:-postgres://postgres:postgres@localhost:5432/competition_host?sslmode=disable}"

clean:
	rm -rf bin



