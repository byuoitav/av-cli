NAME := av-cli
OWNER := byuoitav
PKG := github.com/${OWNER}/${NAME}
PKG_LIST := $(shell go list ${PKG}/...)
# figure out how to get version from somewhere else
VERSION := v1.0.6
BUILD_TIME := $(shell date)
GIT_COMMIT := $(shell git log -1 --pretty="%h")

DIST_BUILD=go build -ldflags "-s -w \
		   -X \"$(PKG)/cmd.version=$(VERSION)\" \
		   -X \"$(PKG)/cmd.buildTime=$(BUILD_TIME)\" \
		   -X \"$(PKG)/cmd.gitCommit=$(GIT_COMMIT)\""

.PHONY: all lint deps test test-cov install dist clean

all:
	go build -o $(NAME) .

lint:
	@golangci-lint run --tests=false

test:
	@go test -v ${PKG_LIST}

test-cov:
	@go test -coverprofile=coverage.txt -covermode=atomic ${PKG_LIST}

deps:
	@go mod download

install:
	go build -o $(NAME) . && mv $(NAME) ${GOPATH}/bin

dist: dist/$(NAME)-linux-amd64 dist/$(NAME)-darwin-amd64 dist/$(NAME)-windows-amd64
	@echo Binaries for version $(VERSION) are located in ./dist/

clean:
	go clean
	rm -rf dist
	rm -f ${GOPATH}/bin/$(NAME)

dist/$(NAME)-linux-amd64:
	env GOOS=linux GOARCH=amd64 $(DIST_BUILD) -o dist/$(NAME)-linux-amd64

dist/$(NAME)-darwin-amd64:
	env GOOS=darwin GOARCH=amd64 $(DIST_BUILD) -o dist/$(NAME)-darwin-amd64

dist/$(NAME)-windows-amd64:
	env GOOS=windows GOARCH=amd64 $(DIST_BUILD) -o dist/$(NAME)-windows-amd64
