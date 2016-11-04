client:
	go run cmd/garrow/cli.go -m client

server:
	go run cmd/garrow/cli.go -m server

debug-client:
	~/workspace/go/src/github.com/golang/go/bin/go run cmd/cli.go -m client

debug-server:
	~/workspace/go/src/github.com/golang/go/bin/go run cmd/cli.go -m client

.PHONY: client server debug-client debug-server
