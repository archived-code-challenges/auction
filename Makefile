#!/usr/bin/make -f

.ONESHELL:
.SHELL := /usr/bin/bash

PROJECTNAME := $(shell basename "$$(pwd)")
PROJECTPATH := $(shell pwd)

help:
	echo "Usage: make [options] [arguments]\n"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' Makefile | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

go-build: ## Compiles packages and dependencies
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go build $(LDFLAGS) "$(PROJECTPATH)/cmd/auction-api/main.go"

go-run: ## Starts API project
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go run $(LDFLAGS) "$(PROJECTPATH)/cmd/auction-api/main.go"

go-doc: ## Generates static docs
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) godoc -http=localhost:6060

docker-build: ## Builds the project binary inside a docker image
	docker build -t $(PROJECTNAME) .

docker-run:	## Runs the previosly build docker image
	docker run $(PROJECTNAME) -p 8080:8080
