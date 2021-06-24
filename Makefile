build:
	go build -o dist/otunnel cmd/main.go
install:
	sudo mv dist/otunnel /usr/local/bin/otunnel
all: build install