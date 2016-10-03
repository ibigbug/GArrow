client:
	go run cmd/cli.go -m client

server:
	go run cmd/cli.go -m server

.PHONY: client server
