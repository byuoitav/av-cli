NAME=av
IMPORT_PATH=github.com/byuoitav/av-cli
VERSION=v1.0.4
BUILD_TIME := $(shell date)
GIT_COMMIT := $(shell git log -1 --pretty="%h")

DIST_BUILD=go build -ldflags "-s -w \
		   -X \"$(IMPORT_PATH)/cmd.version=$(VERSION)\" \
		   -X \"$(IMPORT_PATH)/cmd.buildTime=$(BUILD_TIME)\" \
		   -X \"$(IMPORT_PATH)/cmd.gitCommit=$(GIT_COMMIT)\""

.PHONY: all dist install clean

all:
	go build -o $(NAME) .

install:
	go build -o $(NAME) . && mv $(NAME) ${GOPATH}/bin

dist: dist/$(NAME)-linux-amd64 dist/$(NAME)-darwin-amd64 dist/$(NAME)-windows-amd64
	@echo Binaries for version $(VERSION) are located in ./dist/

dist/$(NAME)-linux-amd64:
	env GOOS=linux GOARCH=amd64 $(DIST_BUILD) -o dist/$(NAME)-linux-amd64

dist/$(NAME)-darwin-amd64:
	env GOOS=darwin GOARCH=amd64 $(DIST_BUILD) -o dist/$(NAME)-darwin-amd64

dist/$(NAME)-windows-amd64:
	env GOOS=windows GOARCH=amd64 $(DIST_BUILD) -o dist/$(NAME)-windows-amd64

clean:
	go clean
	rm -rf dist
	rm -f ${GOPATH}/bin/$(NAME)
