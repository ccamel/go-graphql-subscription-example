.EXPORT_ALL_VARIABLES:

.PHONY: tools deps gen-static check build

GO111MODULE=on

default: build

tools: ./bin/golangci-lint $(GOPATH)/bin/esc $(GOPATH)/bin/goconvey $(GOPATH)/bin/gothanks

deps:
	go get .

gen-static: tools
	go generate main.go

check: tools
	./bin/golangci-lint run ./...

thanks: tools
	$(GOPATH)/bin/gothanks -y | grep -v "is already"

build:
	go build .

goconvey: tools
	$(GOPATH)/bin/goconvey -cover -excludedDirs bin,build,dist,doc,out,etc,vendor

build-linux-amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build .

dockerize:
	docker build -t ccamel/go-graphql-subscription-example .

$(GOPATH)/bin/gothanks:
	@echo "📦 installing $(notdir $@)"
	go get -u github.com/psampaz/gothanks

$(GOPATH)/bin/esc:
	@echo "📦 installing $(notdir $@)"
	go get -u github.com/mjibson/esc

./bin/golangci-lint:
	@echo "📦 installing $(notdir $@)"
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.27.0

$(GOPATH)/bin/goconvey:
	@echo "📦 installing $(notdir $@)"
	go get github.com/smartystreets/goconvey
