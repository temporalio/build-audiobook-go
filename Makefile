# Go parameters
GOCMD = go
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean
GORUN = $(GOCMD) run
GOGET = $(GOCMD) get
GOINSTALL = $(GOCMD) install
GOPATH ?= $(shell go env GOPATH)
BINARY_DIR = bin
BINARY_NAME_WORKER = $(BINARY_DIR)/worker

all: build

clean:
	$(GOCLEAN)
	rm -rf $(BINARY_DIR)

build: $(BINARY_DIR) build-worker

$(BINARY_DIR):
	mkdir -p $(BINARY_DIR)

build-worker:
	cd worker && $(GOBUILD) -o ../$(BINARY_NAME_WORKER)

clean-deps:
	rm -rf $(GOPATH)/pkg/mod

fetch-deps:
	$(GOGET) -u ./...

run:
	./$(BINARY_NAME_WORKER)

.PHONY: all clean build build-worker clean-deps fetch-deps worker
