APP_NAME=stock-manage
MAIN=./cmd/server

.PHONY: run build tidy test

run:
	go run $(MAIN)

build:
	go build -o bin/$(APP_NAME) $(MAIN)

tidy:
	go mod tidy

test:
	go test ./...
