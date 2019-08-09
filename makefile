NAME=av

all:
	go build -o $(NAME) .

install:
	go build -o $(NAME) . && mv $(NAME) ${GOPATH}/bin

linux:
	env GOOS=linux GOARCH=amd64 go build -o $(NAME)

windows:
	env GOOS=windows GOARCH=amd64 go build -o $(NAME).exe

darwin:
	env GOOS=darwin GOARCH=amd64 go build -o $(NAME)

clean:
	go clean
	rm -f $(NAME)
	rm -f $(NAME).exe
	rm -f ${GOPATH}/bin/$(NAME)
