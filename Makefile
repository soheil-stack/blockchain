test:
	go test ./...

build:
	go build -o bin/node cmd/node/main.go

up:
	go run cmd/node/main.go
