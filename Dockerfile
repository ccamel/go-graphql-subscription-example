# Stage build
FROM golang:1.12.7 as builder

WORKDIR /go/src/github.com/ccamel

COPY . .

RUN make build-linux-amd64

# Stage run
FROM scratch

WORKDIR /root/

COPY --from=builder /go/src/github.com/ccamel/go-graphql-subscription-example .

ENTRYPOINT ["./go-graphql-subscription-example"]
