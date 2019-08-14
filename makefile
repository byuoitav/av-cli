NAME=av
IMPORT_PATH=github.com/byuoitav/av-cli
VERSION=v0.9.1

.PHONY: all dist install clean

all:
	go build -o $(NAME) .

install:
	go build -o $(NAME) . && mv $(NAME) ${GOPATH}/bin

dist: dist/$(NAME)-amd64-linux dist/$(NAME)-amd64-darwin dist/$(NAME)-amd64-windows
	@echo Binaries for version $(VERSION) are located in ./dist/

dist/$(NAME)-amd64-linux:
	env GOOS=linux GOARCH=amd64 go build -ldflags "-X $(IMPORT_PATH)/cmd.version=$(VERSION)" -o dist/$(NAME)-amd64-linux

dist/$(NAME)-amd64-darwin:
	env GOOS=darwin GOARCH=amd64 go build -ldflags "-X $(IMPORT_PATH)/cmd.version=$(VERSION)" -o dist/$(NAME)-amd64-darwin

dist/$(NAME)-amd64-windows:
	env GOOS=windows GOARCH=amd64 go build -ldflags "-X $(IMPORT_PATH)/cmd.version=$(VERSION)" -o dist/$(NAME)-amd64-windows

clean:
	go clean
	rm -rf dist
	rm -f $(NAME)-amd64-linux
	rm -f $(NAME)-amd64-darwin
	rm -f $(NAME)-amd64-windows
	rm -f ${GOPATH}/bin/$(NAME)
