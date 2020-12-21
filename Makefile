SHELL := bash
.ONESHELL:
.SHELLFLAGS := -eu -o pipefail -c
.DELETE_ON_ERROR:
MAKEFLAGS += --warn-undefined-variables
MAKEFLAGS += --no-builtin-rules

NAME := go-proxmox
PKG := github.com/markcaudill/$(NAME)

CGO_ENABLED := 0

# Set any default go build tags.
BUILDTAGS :=

# Use > instead of \t for the recipe prefix.
ifeq ($(origin .RECIPEPREFIX), undefined)
  $(error This Make does not support .RECIPEPREFIX. Please use GNU Make 4.0 or later)
endif
.RECIPEPREFIX = >

# Set an output prefix, which is the local directory if not specified
PREFIX?=$(shell pwd)

# Set the build dir, where built cross-compiled binaries will be output
BUILDDIR := ${PREFIX}/cross

# Populate version variables
# Add to compile time flags
VERSION := $(shell cat VERSION.txt)
GITCOMMIT := $(shell git rev-parse --short HEAD)
GITUNTRACKEDCHANGES := $(shell git status --porcelain --untracked-files=no)
ifneq ($(GITUNTRACKEDCHANGES),)
	GITCOMMIT := $(GITCOMMIT)-dirty
endif
ifeq ($(GITCOMMIT),)
	GITCOMMIT := ${GITHUB_SHA}
endif
CTIMEVAR=-X $(PKG)/version.GITCOMMIT=$(GITCOMMIT) -X $(PKG)/version.VERSION=$(VERSION)
GO_LDFLAGS=-ldflags "-w $(CTIMEVAR)"
GO_LDFLAGS_STATIC=-ldflags "-w $(CTIMEVAR) -extldflags -static"

# Set our default go compiler
GO := go

# List the GOOS and GOARCH to build
GOOSARCHES = $(shell cat .goosarch)

# If this session isn't interactive, then we don't want to allocate a
# TTY, which would fail, but if it is interactive, we do want to attach
# so that the user can send e.g. ^C through.
INTERACTIVE := $(shell [ -t 0 ] && echo 1 || echo 0)
ifeq ($(INTERACTIVE), 1)
	DOCKER_FLAGS += -t
endif

help:
>@grep -E '^[^>]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
.PHONY: help

build: prebuild $(NAME) ## Builds a dynamic executable or package.
.PHONY: build

$(NAME): $(wildcard *.go) $(wildcard */*.go) VERSION.txt
>@echo "+ $@"
>$(GO) build -tags "$(BUILDTAGS)" ${GO_LDFLAGS} -o $(NAME) .

static: prebuild ## Builds a static executable.
>@echo "+ $@"
>CGO_ENABLED=$(CGO_ENABLED) $(GO) build \
>			-tags "$(BUILDTAGS) static_build" \
>			${GO_LDFLAGS_STATIC} -o $(NAME) .
.PHONY: static

all: clean build fmt lint test vet install ## Runs a clean, build, fmt, lint, test, staticcheck, vet and install.

fmt: ## Verifies all files have been `gofmt`ed.
>@echo "+ $@"
>@if [[ ! -z "$(shell gofmt -s -l . | grep -v '.pb.go:' | grep -v '.twirp.go:' | grep -v vendor | tee /dev/stderr)" ]]; then \
>	exit 1; \
>fi

.PHONY: lint
lint: ## Verifies `golint` passes.
>@echo "+ $@"
>@if [[ ! -z "$(shell golint ./... | grep -v '.pb.go:' | grep -v '.twirp.go:' | grep -v vendor | tee /dev/stderr)" ]]; then \
>	exit 1; \
>fi
.PHONY: fmt

test: prebuild ## Runs the go tests.
>@echo "+ $@"
>@$(GO) test -v -tags "$(BUILDTAGS) cgo" $(shell $(GO) list ./... | grep -v vendor)
.PHONY: test

vet: ## Verifies `go vet` passes.
>@echo "+ $@"
>@if [[ ! -z "$(shell $(GO) vet $(shell $(GO) list ./... | grep -v vendor) | tee /dev/stderr)" ]]; then \
>	exit 1; \
>fi
.PHONY: vet

cover: prebuild ## Runs go test with coverage.
>@echo "" > coverage.txt
>@for d in $(shell $(GO) list ./... | grep -v vendor); do \
>	$(GO) test -race -coverprofile=profile.out -covermode=atomic "$$d"; \
>	if [ -f profile.out ]; then \
>		cat profile.out >> coverage.txt; \
>		rm profile.out; \
>	fi; \
>done;
.PHONY: cover

install: prebuild ## Installs the executable or package.
>@echo "+ $@"
>$(GO) install -a -tags "$(BUILDTAGS)" ${GO_LDFLAGS} .
.PHONY: install

bump-major-version: ## Bump the major version in the version file.
>$(eval NEW_VERSION = $(shell echo $(VERSION) | sed 's/^[a-zA-Z]*//g' | awk -F'.' '{print ($$1 + 1)"."$$2"."$$3}'))
>@echo v$(NEW_VERSION) > VERSION.txt
>@sed -i s/$(VERSION)/$(NEW_VERSION)/g README.md
>git add VERSION.txt README.md
>git commit -vsam "Bump version to $(NEW_VERSION)"
>@echo "Run make tag to create and push the tag for new version $(NEW_VERSION)"
.PHONY: bump-major-version

bump-minor-version: ## Bump the minor version in the version file.
>$(eval NEW_VERSION = $(shell echo $(VERSION) | sed 's/^[a-zA-Z]*//g' | awk -F'.' '{print $$1"."($$2 + 1)"."$$3}'))
>@echo v$(NEW_VERSION) > VERSION.txt
>@sed -i s/$(VERSION)/$(NEW_VERSION)/g README.md
>git add VERSION.txt README.md
>git commit -vsam "Bump version to $(NEW_VERSION)"
>@echo "Run make tag to create and push the tag for new version $(NEW_VERSION)"
.PHONY: bump-minor-version

bump-patch-version: ## Bump the patch version in the version file.
>$(eval NEW_VERSION = $(shell echo $(VERSION) | sed 's/^[a-zA-Z]*//g' | awk -F'.' '{print $$1"."$$2"."($$3 + 1)}'))
>@echo v$(NEW_VERSION) > VERSION.txt
>@sed -i s/$(VERSION)/$(NEW_VERSION)/g README.md
>git add VERSION.txt README.md
>git commit -vsam "Bump version to $(NEW_VERSION)"
>@echo "Run make tag to create and push the tag for new version $(NEW_VERSION)"
.PHONY: bump-patch-version

tag: ## Create a new git tag to prepare to build a release.
>git tag -sa $(VERSION) -m "$(VERSION)"
>@echo "Run git push origin $(VERSION) to push your new tag to GitHub and trigger a release."
.PHONY: tag

AUTHORS:
>@$(file >$@,# This file lists all individuals having contributed content to the repository.)
>@$(file >>$@,# For how it is generated, see `make AUTHORS`.)
>@echo "$(shell git log --format='%aN <%aE>' | LC_ALL=C.UTF-8 sort -uf)" >> $@
.PHONY: AUTHORS

vendor: ## Updates the vendoring directory.
>@$(RM) go.sum
>@$(RM) -r vendor
>GO111MODULE=on $(GO) mod init || true
>GO111MODULE=on $(GO) mod tidy
>GO111MODULE=on $(GO) mod vendor
>@$(RM) Gopkg.toml Gopkg.lock
.PHONY: vendor

clean: ## Cleanup any build binaries or packages.
>@echo "+ $@"
>$(RM) $(NAME)
>$(RM) -r $(BUILDDIR)
.PHONY: clean

prebuild:
.PHONY: prebuild
