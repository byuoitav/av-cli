NAME=av

all:
	go build -o $(NAME) .

install:
	go build -o $(NAME) . && mv $(NAME) ${GOPATH}/bin

linux:
	env GOOS=linux GOARCH=amd64 go build -o $(NAME)

windows:
	env GOOS=windows GOARCH=amd64 go build -o $(NAME)

mac:
	env GOOS=darwin GOARCH=amd64 go build -o $(NAME)

clean:
	go clean
	rm -f $(NAME)
	rm -f ${GOPATH}/bin/$(NAME)
