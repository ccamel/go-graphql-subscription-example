package server

import (
	"context"

	graphql "github.com/ccamel/go-graphql-subscription-example/server/scalar"
)

type resolver struct {
}

func newResolver() *resolver {
	return &resolver{}
}

func (r *resolver) Event(ctx context.Context) <-chan *graphql.JSONObject {
	c := make(chan *graphql.JSONObject)

	go consume(ctx, c)

	return c
}

func (r *resolver) Foo() *string {
	v := "ok"

	return &v
}
