.PHONY: build test lint clean

build:
	go build -o bin/cxas ./cmd/cxas

test:
	go test ./... -v -race

test-integration:
	go test -tags=integration ./... -v

lint:
	golangci-lint run ./...

clean:
	rm -rf bin/
