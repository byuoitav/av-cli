NAME := av-cli
OWNER := byuoitav
PKG := github.com/${OWNER}/${NAME}
DOCKER_URL := docker.pkg.github.com
DOCKER_PKG := ${DOCKER_URL}/${OWNER}/${NAME}

# version:
# use the git tag, if this commit
# doesn't have a tag, use the git hash
COMMIT_HASH := $(shell git rev-parse --short HEAD)
TAG := $(shell git rev-parse --short HEAD)
ifneq ($(shell git describe --exact-match --tags HEAD 2> /dev/null),)
	TAG = $(shell git describe --exact-match --tags HEAD)
endif

#BUILD_TIME := $(shell date)
#COMMIT_HASH := $(shell git rev-parse --short HEAD)
#VERSION := $(shell git rev-parse --short HEAD)
#ifneq ($(shell git describe --exact-match --tags HEAD 2> /dev/null),)
#	VERSION = $(shell git describe --exact-match --tags HEAD)
#endif

PRD_TAG_REGEX := "v[0-9]+\.[0-9]+\.[0-9]+"
DEV_TAG_REGEX := "v[0-9]+\.[0-9]+\.[0-9]+-.+"

# go stuff
PKG_LIST := $(shell go list ${PKG}/...)

#BUILD_CLI=go build -ldflags "-s -w \
		   -X \"$(CLI_PKG)/cmd.version=$(VERSION)\" \
		   -X \"$(CLI_PKG)/cmd.buildTime=$(BUILD_TIME)\" \
		   -X \"$(CLI_PKG)/cmd.gitCommit=$(COMMIT_HASH)\""

.PHONY: all deps build test test-cov clean

all: clean build

test:
	@go test -v ${PKG_LIST}

test-cov:
	@go test -coverprofile=coverage.txt -covermode=atomic ${PKG_LIST}

lint:
	@golangci-lint run --tests=false

# must have protoc installed
deps:
	@echo Generating protobuf files...
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@go generate ./...

	@echo Downloading dependencies...
	@go mod download

#install:
#	go build -o $(NAME) . && mv $(NAME) ${GOPATH}/bin

build: deps
	@mkdir -p dist

	@echo
	@echo Building api for linux-amd64
	@cd cmd/api/ && env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o ../../dist/api-linux-amd64

	@echo
	@echo Building slack for linux-amd64
	@cd cmd/slack/ && env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o ../../dist/slack-linux-amd64

	@echo
	@echo Build output is located in ./dist/.

#	@echo
#	@echo Building CLI for linux-amd64
#	@cd cli && env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(BUILD_CLI) -o ../dist/${NAME}-cli-linux-amd64 ${CLI_PKG}
#
#	@echo
#	@echo Building CLI for darwin-amd64
#	@cd cli && env CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(BUILD_CLI) -o ../dist/${NAME}-cli-darwin-amd64 ${CLI_PKG}
#
#	@echo
#	@echo Building CLI for windows-amd64
#	@cd cli && env CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(BUILD_CLI) -o ../dist/${NAME}-cli-windows-amd64 ${CLI_PKG}

docker: clean build
ifeq (${COMMIT_HASH}, ${TAG})
	@echo Building dev containers with tag ${COMMIT_HASH}

	@echo Building container ${DOCKER_PKG}/api-dev:${COMMIT_HASH}
	@docker build -f dockerfile --build-arg NAME=api-linux-amd64 -t ${DOCKER_PKG}/api-dev:${COMMIT_HASH} dist

	@echo Building container ${DOCKER_PKG}/slack-dev:${COMMIT_HASH}
	@docker build -f dockerfile --build-arg NAME=slack-linux-amd64 -t ${DOCKER_PKG}/slack-dev:${COMMIT_HASH} dist
else ifneq ($(shell echo ${TAG} | grep -x -E ${DEV_TAG_REGEX}),)
	@echo Building dev containers with tag ${TAG}

	@echo Building container ${DOCKER_PKG}/api-dev:${TAG}
	@docker build -f dockerfile --build-arg NAME=api-linux-amd64 -t ${DOCKER_PKG}/api-dev:${TAG} dist

	@echo Building container ${DOCKER_PKG}/slack-dev:${TAG}
	@docker build -f dockerfile --build-arg NAME=slack-linux-amd64 -t ${DOCKER_PKG}/slack-dev:${TAG} dist
else ifneq ($(shell echo ${TAG} | grep -x -E ${PRD_TAG_REGEX}),)
	@echo Building prd containers with tag ${TAG}

	@echo Building container ${DOCKER_PKG}/api:${TAG}
	@docker build -f dockerfile --build-arg NAME=api-linux-amd64 -t ${DOCKER_PKG}/api:${TAG} dist

	@echo Building container ${DOCKER_PKG}/slack:${TAG}
	@docker build -f dockerfile --build-arg NAME=slack-linux-amd64 -t ${DOCKER_PKG}/slack:${TAG} dist
endif

deploy: docker
	@echo Logging into Github Package Registry
	@docker login ${DOCKER_URL} -u ${DOCKER_USERNAME} -p ${DOCKER_PASSWORD}

ifeq (${COMMIT_HASH}, ${TAG})
	@echo Pushing dev containers with tag ${COMMIT_HASH}

	@echo Pushing container ${DOCKER_PKG}/api-dev:${COMMIT_HASH}
	@docker push ${DOCKER_PKG}/api-dev:${COMMIT_HASH}

	@echo Pushing container ${DOCKER_PKG}/slack-dev:${COMMIT_HASH}
	@docker push ${DOCKER_PKG}/slack-dev:${COMMIT_HASH}
else ifneq ($(shell echo ${TAG} | grep -x -E ${DEV_TAG_REGEX}),)
	@echo Pushing dev containers with tag ${TAG}

	@echo Pushing container ${DOCKER_PKG}/api-dev:${TAG}
	@docker push ${DOCKER_PKG}/api-dev:${TAG}

	@echo Pushing container ${DOCKER_PKG}/slack-dev:${TAG}
	@docker push ${DOCKER_PKG}/slack-dev:${TAG}
else ifneq ($(shell echo ${TAG} | grep -x -E ${PRD_TAG_REGEX}),)
	@echo Pushing prd containers with tag ${TAG}

	@echo Pushing container ${DOCKER_PKG}/api:${TAG}
	@docker push ${DOCKER_PKG}/api:${TAG}

	@echo Pushing container ${DOCKER_PKG}/slack:${TAG}
	@docker push ${DOCKER_PKG}/slack:${TAG}
endif

clean:
	@go clean
	@rm -rf dist/
