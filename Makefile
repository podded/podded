# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
BINARY_NAME=podded

all: test build

build:
	$(GOBUILD) -o bin/$(BINARY_NAME) cmd/main.go

test:
	$(GOTEST) -v ./...

lint:
	golangci-lint run

clean:
	$(GOCLEAN)
	rm -rf bin/
