all:
	go build -o av .

install:
	go build -o av . && mv av ${GOPATH}/bin
