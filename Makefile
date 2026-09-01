.PHONY: all build test clean run fmt vet

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOFMT=gofmt
GOVET=$(GOCMD) vet

# Target binaries
NODE_BINARY=bin/node
CLI_BINARY=bin/cli

all: build test

build:
	mkdir -p bin
	$(GOBUILD) -o $(NODE_BINARY) ./cmd/node/main.go
	$(GOBUILD) -o $(CLI_BINARY) ./cmd/cli/main.go

run: build
	./$(NODE_BINARY) --config configs/config.yaml

run-a: build
	./$(NODE_BINARY) --config configs/node_a.yaml

run-b: build
	./$(NODE_BINARY) --config configs/node_b.yaml

test:
	$(GOTEST) -v -race ./...

fmt:
	$(GOFMT) -s -w .

vet:
	$(GOVET) ./...

clean:
	$(GOCLEAN)
	rm -rf bin/
	rm -rf data/*
