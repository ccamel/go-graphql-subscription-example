.EXPORT_ALL_VARIABLES:

.PHONY: install-tools install-deps gen-static check build

GO111MODULE=on

default: build

install-tools:
	@if [ ! -f $(GOPATH)/bin/esc ]; then \
		echo "installing esc..."; \
		go get -u github.com/mjibson/esc; \
	fi
	@if [ ! -f ./bin/golangci-lint ]; then \
		echo "installing golangci-lint..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.27.0; \
	fi
	@if [ ! -f $(GOPATH)/bin/gothanks ]; then \
		echo "installing gothanks..."; \
		go get -u github.com/psampaz/gothanks; \
	fi

install-deps:
	go get .

gen-static: install-tools
	go generate main.go

check: install-tools
	./bin/golangci-lint run ./...

thanks: install-tools
	$(GOPATH)/bin/gothanks -y | grep -v "is already"

build:
	go build .

build-linux-amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build .

dockerize:
	docker build -t ccamel/go-graphql-subscription-example .
