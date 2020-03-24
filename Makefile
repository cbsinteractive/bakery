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
clean: 
				$(GOCLEAN) ./...
				rm -f $(BINARY_NAME)
run:
				$(GORUN) $(WEBSERVER)

