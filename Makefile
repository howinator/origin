.PHONY: build install test lint clean

build:
	go build -o bin/house ./cmd/house

install:
	go install ./cmd/house

test:
	go test ./...

lint:
	golangci-lint run ./...

clean:
	rm -rf bin/
