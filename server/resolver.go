package server

import (
	"context"
	"fmt"

	uuid "github.com/satori/go.uuid"

	"github.com/rs/zerolog"

	graphql "github.com/ccamel/go-graphql-subscription-example/server/scalar"
)

type ctxKey int

const (
	subscriptionID ctxKey = iota
	brokersKey
	topicKey
	offsetKey
)

type Resolver struct {
	log zerolog.Logger
	cfg *Configuration
}

func NewResolver(cfg *Configuration, log zerolog.Logger) *Resolver {
	return &Resolver{log, cfg}
}

func (r *Resolver) Event(
	ctx context.Context,
	args *struct {
		On string
		At graphql.Offset
	}) (<-chan *graphql.JSONObject, error) {
	if !acceptTopic(args.On, r.cfg.Topics) {
		return nil, fmt.Errorf("unknown topic: '%s'. Valid topics are: %v", args.On, r.cfg.Topics)
	}
	c := make(chan *graphql.JSONObject)

	ctx = context.WithValue(ctx, subscriptionID, uuid.NewV4().String())
	ctx = context.WithValue(ctx, brokersKey, r.cfg.Brokers)

	ctx = context.WithValue(ctx, topicKey, args.On)
	ctx = context.WithValue(ctx, offsetKey, args.At)

	ctx = r.log.WithContext(ctx)

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
