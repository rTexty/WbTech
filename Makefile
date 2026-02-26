.PHONY: run producer test lint tidy

run:
	go run cmd/server/main.go

producer:
	go run cmd/producer/main.go

test:
	go test ./...

lint:
	golangci-lint run ./cmd/... ./internal/...

tidy:
	go mod tidy && go fmt ./...