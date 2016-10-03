test:
	go test -race -v ./...

ci:
	go test -v ./... -covermode=count -coverprofile=coverage.out
	$(HOME)/gopath/bin/goveralls -coverprofile=coverage.out -service=travis-ci -repotoken $(COVERALLS_TOKEN)

.PHONY: test
