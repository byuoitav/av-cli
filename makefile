NAME=av

all:
	go build -o $(NAME) .

install:
	go build -o $(NAME) . && mv $(NAME) ${GOPATH}/bin

clean:
	go clean
	rm -f $(NAME)
	rm -f ${GOPATH}/bin/$(NAME)
