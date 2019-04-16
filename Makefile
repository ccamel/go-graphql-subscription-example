.EXPORT_ALL_VARIABLES:

GO111MODULE=on

default: build

install-tools:
	@if [ ! -f $(GOPATH)/bin/esc ]; then \
		echo "installing esc..."; \
		go get -u github.com/mjibson/esc; \
	fi

gen-static: install-tools
	go generate main.go

check:
	golangci-lint run ./...

build: gen-static
	go build .