SHELL = /bin/bash
### https://makefiletutorial.com/

-include .env
export


##@ Help
help:  ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-25s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Build

build: ## Build the binary
	go build -o myip ./cmd/myip

##@ Run

run: build ## Build and run the binary
	./myip

logs: ## View application logs (macOS)
	log show --predicate 'process == "myip"' --last 1h

##@ Test

test: ## Run tests
	go test -v ./...

test-coverage: ## Run tests with coverage
	go test -v -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

goreleaser: ## Install goreleaser
	go install github.com/goreleaser/goreleaser/v2@latest

test-release: ## Test releaser
	goreleaser release --snapshot --clean --skip=publish

lint: ## Run linter
	go vet ./...

##@ Docker

keydb: ## Run KeyDB in docker
	docker run -d --name keydb -p 6378:6379 -v myip-keydb:/data eqalpha/keydb

keydb-stop: ## Stop KeyDB docker container
	docker stop keydb && docker rm keydb
	