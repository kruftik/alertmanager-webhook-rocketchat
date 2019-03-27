
SRC_DIR=github.com/fxinnovation/alertmanager-webhook-rocketchat
BUILD_VERSION=$(shell cat VERSION.txt)
APPL=alertmanager-webhook-rocketchat

######## commom

PKGGOFILES=$(shell go list ./... | grep -v /vendor/)

GIT_COMMIT?=$(shell git rev-parse --short HEAD)
GIT_DIRTY?=$(shell test -n "`git status --porcelain`" && echo "+CHANGES" || true)
GIT_DESCRIBE?=$(shell git describe --tags --always)
BUILD_TIME?=$(shell date +"%Y-%m-%dT%H:%M:%S")

LDFLAGS=-ldflags "\
          -X $(SRC_DIR)/information.Version=$(BUILD_VERSION) \
          -X $(SRC_DIR)/information.BuildTime=$(BUILD_TIME) \
          -X $(SRC_DIR)/information.GitCommit=$(GIT_COMMIT) \
          -X $(SRC_DIR)/information.GitDirty=$(GIT_DIRTY) \
          -X $(SRC_DIR)/information.GitDescribe=$(GIT_DESCRIBE)"


PWD=$(shell pwd)

.PHONY: help
help:
	@grep -hE '^[a-zA-Z_-]+.*?:.*?## .*$$' ${MAKEFILE_LIST} | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[0;49;95m%-30s\033[0m %s\n", $$1, $$2}'

## If you have go on your wonderful laptop
.PHONY: clean
clean:
	@rm -rf ./target || true
	@mkdir ./target || true

.PHONY: test
test: fmt vet ## go test
	go test -cpu=2 -p=2 -v --short $(LDFLAGS) $(PKGGOFILES)

.PHONY: test-it-test
test-it-test: fmt vet ## go test with integration
	go test $(PKGGOFILES) -cpu=2 -p=2 -race  -v $(LDFLAGS)

.PHONY: test-cover
test-cover: fmt vet ## go test with coverage
	go test  $(PKGGOFILES) -cover -race -v $(LDFLAGS)

.PHONY: test-coverage
test-coverage: clean fmt vet ## for jenkins
	gocov test $(PKGGOFILES) --short -cpu=2 -p=2 -v $(LDFLAGS) | gocov-xml > ./coverage-test.xml

.PHONY: test-it-test-coverage
test-it-test-coverage: clean fmt vet ## for jenkins
	gocov test $(PKGGOFILES) -cpu=2 -p=2 -v $(LDFLAGS) | gocov-xml > ./coverage-test-it-test.xml

.PHONY: dependencies
dependencies: ## download the dependencies
	rm -rf Gopkg.lock vendor/
	dep ensure

.PHONY: build
build: clean fmt vet
	go build $(LDFLAGS)

.PHONY: fmt
fmt: ## go fmt on packages
	go fmt $(PKGGOFILES)

.PHONY: vet
vet: ## go vet on packages
	go vet $(PKGGOFILES)

.PHONY: lint
lint: ## go vet on packages
	golint -set_exit_status=true $(PKGGOFILES)

.PHONY: tools
tools: ## install tools to develop
	go get -u github.com/golang/dep/cmd/dep
	go get -u golang.org/x/lint/golint
	go get github.com/axw/gocov/...
	go get github.com/AlekSi/gocov-xml
