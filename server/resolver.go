package server

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"

	graphql "github.com/ccamel/go-graphql-subscription-example/server/scalar"
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
		On       string
		At       graphql.Offset
		Matching *string
	}) (<-chan *graphql.JSONObject, error) {
	if !acceptTopic(args.On, r.cfg.Topics) {
		return nil, fmt.Errorf("unknown topic: '%s'. Valid topics are: %v", args.On, r.cfg.Topics)
	}
	c := make(chan *graphql.JSONObject)

	ctx = r.log.WithContext(ctx)

	consumer, err := NewConsumer(
		ctx,
		r.cfg.Brokers,
		args.On,
		args.At.Value().Int64(),
		args.Matching,
		c,
	)
	if err != nil {
		return nil, err
	}

	go consumer.Start()

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
