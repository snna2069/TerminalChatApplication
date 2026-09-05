.PHONY: run-server run-client test test-race vet build clean

run-server:
	go run ./cmd/server -addr localhost:8080 -data data/messages.jsonl

run-client:
	go run ./cmd/client -addr localhost:8080

test:
	go test ./...

test-race:
	go test -race ./...

vet:
	go vet ./...

build:
	go build -o bin/server ./cmd/server
	go build -o bin/client ./cmd/client

clean:
	go clean
	Remove-Item -Recurse -Force bin -ErrorAction SilentlyContinue
