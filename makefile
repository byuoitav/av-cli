NAME := av-cli
OWNER := byuoitav
API_PKG := github.com/${OWNER}/${NAME}
CLI_PKG := github.com/${OWNER}/${NAME}/cli
SLACK_PKG := github.com/${OWNER}/${NAME}/slack
DOCKER_URL := docker.pkg.github.com

BUILD_TIME := $(shell date)
COMMIT_HASH := $(shell git rev-parse --short HEAD)
VERSION := $(shell git rev-parse --short HEAD)
ifneq ($(shell git describe --exact-match --tags HEAD 2> /dev/null),)
	VERSION = $(shell git describe --exact-match --tags HEAD)
endif

API_PKG_LIST := $(shell go list ${API_PKG}/...)
CLI_PKG_LIST := $(shell cd cli && go list ${CLI_PKG}/...)
SLACK_PKG_LIST := $(shell cd slack && go list ${SLACK_PKG}/...)

BUILD_CLI=go build -ldflags "-s -w \
		   -X \"$(PKG)/cmd.version=$(VERSION)\" \
		   -X \"$(PKG)/cmd.buildTime=$(BUILD_TIME)\" \
		   -X \"$(PKG)/cmd.gitCommit=$(COMMIT_HASH)\""

.PHONY: all deps build test test-cov clean lint install

all: clean build

lint:
	@echo Linting api
	@golangci-lint run --tests=false

	@echo Linting cli
	@cd cli && golangci-lint run --tests=false

	@echo Linting slack
	@cd slack && golangci-lint run --tests=false

test:
	@echo Testing api
	@go test -v ${API_PKG_LIST}

	@echo Testing cli
	@cd cli && go test -v ${CLI_PKG_LIST}

	@echo Testing slack
	@cd slack && go test -v ${SLACK_PKG_LIST}

#test-cov:
#	@cd api && go test -coverprofile=coverage.txt -covermode=atomic ${API_PKG_LIST}
#	@cd cli && go test -coverprofile=coverage.txt -covermode=atomic ${CLI_PKG_LIST}
#	@cd slack && go test -coverprofile=coverage.txt -covermode=atomic ${SLACK_PKG_LIST}

# must have protoc installed
deps:
	@echo Generating protobuf files...
	@go get -u github.com/golang/protobuf/protoc-gen-go
	@go generate ./...

	@go mod download
	@cd cli && go mod download
	@cd slack && go mod download

#install:
#	go build -o $(NAME) . && mv $(NAME) ${GOPATH}/bin

build: deps
	@mkdir -p dist

	@echo
	@echo Building API for linux-amd64
	@env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o ./dist/${NAME}-api-linux-amd64 ${API_PKG}/api

	@echo
	@echo Building CLI for linux-amd64
	@cd cli && env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(BUILD_CLI) -o ../dist/${NAME}-cli-linux-amd64 ${CLI_PKG}

	@echo
	@echo Building CLI for darwin-amd64
	@cd cli && env CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(BUILD_CLI) -o ../dist/${NAME}-cli-darwin-amd64 ${CLI_PKG}

	@echo
	@echo Building CLI for windows-amd64
	@cd cli && env CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(BUILD_CLI) -o ../dist/${NAME}-cli-windows-amd64 ${CLI_PKG}

	@echo
	@echo Building slackbot for linux-amd64
	@cd slack && env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o ../dist/${NAME}-slack-linux-amd64 ${SLACK_PKG}

	@echo
	@echo Build output is located in ./dist/.

docker: clean build
ifneq (${COMMIT_HASH},${VERSION})
	@echo Building container ${DOCKER_URL}/${OWNER}/${NAME}/api:${VERSION}
	@docker build -f dockerfile --build-arg NAME=${NAME}-api-linux-amd64 -t ${DOCKER_URL}/${OWNER}/${NAME}/api:${VERSION} dist

	@echo Building container ${DOCKER_URL}/${OWNER}/${NAME}/slack:${VERSION}
	@docker build -f dockerfile --build-arg NAME=${NAME}-slack-linux-amd64 -t ${DOCKER_URL}/${OWNER}/${NAME}/slack:${VERSION} dist
else
	@echo Building container ${DOCKER_URL}/${OWNER}/${NAME}/api-dev:${COMMIT_HASH}
	@docker build -f dockerfile --build-arg NAME=${NAME}-api-linux-amd64 -t ${DOCKER_URL}/${OWNER}/${NAME}/api-dev:${COMMIT_HASH} dist

	@echo Building container ${DOCKER_URL}/${OWNER}/${NAME}/slack-dev:${COMMIT_HASH}
	@docker build -f dockerfile --build-arg NAME=${NAME}-slack-linux-amd64 -t ${DOCKER_URL}/${OWNER}/${NAME}/slack-dev:${COMMIT_HASH} dist
endif

deploy: docker
	@echo Logging into ${DOCKER_URL}
	@docker login ${DOCKER_URL} -u ${DOCKER_USERNAME} -p ${DOCKER_PASSWORD}

ifneq (${COMMIT_HASH},${VERSION})
	@echo Pushing container ${DOCKER_URL}/${OWNER}/${NAME}/api:${VERSION}
	@docker push ${DOCKER_URL}/${OWNER}/${NAME}/api:${VERSION}

	@echo Pushing container ${DOCKER_URL}/${OWNER}/${NAME}/slack:${VERSION}
	@docker push ${DOCKER_URL}/${OWNER}/${NAME}/slack:${VERSION}
else
	@echo Pushing container ${DOCKER_URL}/${OWNER}/${NAME}/api-dev:${COMMIT_HASH}
	@docker push ${DOCKER_URL}/${OWNER}/${NAME}/api-dev:${COMMIT_HASH}

	@echo Pushing container ${DOCKER_URL}/${OWNER}/${NAME}/slack-dev:${COMMIT_HASH}
	@docker push ${DOCKER_URL}/${OWNER}/${NAME}/slack-dev:${COMMIT_HASH}
endif

clean:
	@cd api && go clean
	@cd cli && go clean
	@cd slack && go clean
	rm -f ${GOPATH}/bin/$(NAME)
	rm -rf dist
