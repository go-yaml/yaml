.PHONY: test

install-libs:
	go mod vendor

build:
	go build src/main.go

test:
	go test .
