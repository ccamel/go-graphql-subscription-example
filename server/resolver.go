package server

import (
	"context"
	"fmt"

	graphql "github.com/ccamel/go-graphql-subscription-example/server/scalar"
)

type resolver struct {
	cfg *Configuration
}

func NewResolver(cfg *Configuration) *resolver {
	return &resolver{
		cfg,
	}
}

func (r *resolver) Event(
	ctx context.Context,
	args *struct {
		Topic string
	}) (<-chan *graphql.JSONObject, error) {
	if !acceptTopic(args.Topic, r.cfg.Topics) {
		return nil, fmt.Errorf("Unknown topic: '%s'. Valid topics are: %v", r.cfg.Topics)
	}

	c := make(chan *graphql.JSONObject)

	ctx = context.WithValue(ctx, "topic", args.Topic)
	ctx = context.WithValue(ctx, "brokers", r.cfg.Brokers)

	go consume(ctx, c)

	return c, nil
}

func (r *resolver) Topics() []string {
	return r.cfg.Topics
}

func acceptTopic(topic string, topics []string) bool {
	for _, t := range topics {
		if t == topic {
			return true
		}
	}
	return false
}
