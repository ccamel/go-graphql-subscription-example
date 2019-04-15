.EXPORT_ALL_VARIABLES:

GO111MODULE=on

default: build

gen-static:
	go generate main.go

check:
	golangci-lint run ./...

build: gen-static
	go build .