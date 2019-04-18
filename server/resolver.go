package server

import (
	"context"
	"fmt"

	graphql "github.com/ccamel/go-graphql-subscription-example/server/scalar"
)

type ctxKey int

const (
	topicKey ctxKey = iota
	brokersKey
)

type Resolver struct {
	cfg *Configuration
}

func NewResolver(cfg *Configuration) *Resolver {
	return &Resolver{
		cfg,
	}
}

func (r *Resolver) Event(
	ctx context.Context,
	args *struct {
		Topic string
	}) (<-chan *graphql.JSONObject, error) {
	if !acceptTopic(args.Topic, r.cfg.Topics) {
		return nil, fmt.Errorf("unknown topic: '%s'. Valid topics are: %v", args.Topic, r.cfg.Topics)
	}

	c := make(chan *graphql.JSONObject)

	ctx = context.WithValue(ctx, topicKey, args.Topic)
	ctx = context.WithValue(ctx, brokersKey, r.cfg.Brokers)

	go consume(ctx, c)

	return c, nil
}

func (r *Resolver) Topics() []string {
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
