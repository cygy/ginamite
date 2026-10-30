GO       ?= go
GIT      ?= git
REMOTE   ?= origin
VERSION  := $(shell cat VERSION)
COVERAGE ?= coverage.out

.DEFAULT_GOAL := help

.PHONY: help fmt fmt-check vet lint test cover build tidy check clean \
	bump_major bump_minor bump_patch release ensure-clean

help: ## List the targets
	@awk 'BEGIN {FS = ":.*## "} /^[a-z_-]+:.*## / {printf "  %-12s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

fmt: ## Format the code with gofumpt and goimports
	gofumpt -w .
	goimports -w .

fmt-check: ## Fail if a file is not formatted
	@out="$$(gofumpt -l .)"; \
	if [ -n "$$out" ]; then echo "not formatted:"; echo "$$out"; exit 1; fi

vet: ## Run go vet
	$(GO) vet ./...

lint: ## Run golangci-lint
	golangci-lint run ./...

test: ## Run the tests with the race detector
	$(GO) test -race -count=1 ./...

cover: ## Run the tests and print the coverage per function
	$(GO) test -race -count=1 -coverprofile=$(COVERAGE) ./...
	$(GO) tool cover -func=$(COVERAGE) | tail -n 1

build: ## Compile every package
	$(GO) build ./...

tidy: ## Tidy the go.mod and go.sum files
	$(GO) mod tidy

check: build fmt-check vet lint test ## Run every verification

clean: ## Remove the generated files
	rm -f $(COVERAGE)

bump_major: ## Release the next major version
	@$(MAKE) release NEXT="$(shell echo $(VERSION) | awk -F. '{print $$1+1".0.0"}')"

bump_minor: ## Release the next minor version
	@$(MAKE) release NEXT="$(shell echo $(VERSION) | awk -F. '{print $$1"."$$2+1".0"}')"

bump_patch: ## Release the next patch version
	@$(MAKE) release NEXT="$(shell echo $(VERSION) | awk -F. '{print $$1"."$$2"."$$3+1}')"

ensure-clean:
	@test -z "$$($(GIT) status --porcelain)" || { echo "the working tree is not clean"; exit 1; }

# release is internal: use bump_major, bump_minor or bump_patch.
release: ensure-clean check
	@test -n "$(NEXT)" || { echo "NEXT is not set"; exit 1; }
	echo $(NEXT) > VERSION
	$(GO) mod tidy
	$(GIT) add VERSION go.mod go.sum
	$(GIT) commit -s -m "Version v$(NEXT)"
	$(GIT) tag -a v$(NEXT) -m "Version v$(NEXT)"
	$(GIT) push $(REMOTE)
	$(GIT) push $(REMOTE) v$(NEXT)
