ifeq ($(shell git tag --contains HEAD),)
  VERSION := $(shell git rev-parse --short HEAD)
else
  VERSION := $(shell git tag --contains HEAD)
endif
BUILDTIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
GOLDFLAGS += -X github.com/timsolov/fragmented-tcp/server/server.Version=$(VERSION)
GOLDFLAGS += -X github.com/timsolov/fragmented-tcp/server/server.Buildtime=$(BUILDTIME)
GOFLAGS = -ldflags "$(GOLDFLAGS)"

NAME := fragmented-tcp
COMPOSE_FILE_PATH := build/docker-compose.yaml


.PHONY: help build gen test

.DEFAULT_GOAL := build

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: test ## Build application (default goal)
	GOSUMDB=off \
	go build -o $(NAME)-server $(GOFLAGS) -v ./cmd/server

test: ## Run all tests
	go test ./...

gen: ## Perform go generate all
	go generate ./...

