# Stage build
FROM golang:1.24.2 as builder

WORKDIR /tmp

RUN    git clone --depth 1 -b v1.15.0 https://github.com/magefile/mage.git \
    && cd mage \
    && go run bootstrap.go install

WORKDIR /go/src/github.com/ccamel

COPY . .

RUN mage linux_amd64_build

# Stage run
FROM scratch

WORKDIR /root/

COPY --from=builder /go/src/github.com/ccamel/go-graphql-subscription-example .

ENTRYPOINT ["./go-graphql-subscription-example"]
