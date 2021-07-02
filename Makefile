# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GORUN=$(GOCMD) run
GOTEST=$(GOCMD) test
BINARY_NAME=bakery
WEBSERVER=./cmd/http

all: test build

build: 
	$(GOBUILD) -o $(BINARY_NAME) -v $(WEBSERVER)

test: 
	$(GOTEST) -v -race -count=1 ./...

test_cover:
	GOFLAGS=-p=8 $(GOTEST) -v -count 1 ./... -race -coverprofile=coverage.txt -covermode=atomic

clean: 
	$(GOCLEAN) ./...
	rm -f $(BINARY_NAME)

run:
	$(GORUN) $(WEBSERVER)
