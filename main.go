package main

import "github.com/ccamel/go-graphql-subscription-example/server"

//go:generate esc -o static/static.go -pkg static static

func main() {
	server.StartServer()
}
