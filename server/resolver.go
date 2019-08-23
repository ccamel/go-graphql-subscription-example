package server

import (
	"context"
	"fmt"
	"net/url"

	"github.com/antonmedv/expr"
	"github.com/reactivex/rxgo/handlers"
	"github.com/reactivex/rxgo/observer"

	"github.com/rs/zerolog"

	"github.com/ccamel/go-graphql-subscription-example/server/scalar"
)

type Resolver struct {
	log zerolog.Logger
	cfg *Configuration

	source Source
}

func NewResolver(cfg *Configuration, log zerolog.Logger) (*Resolver, error) {
	sourceURI, err := url.Parse(cfg.Source)
	if err != nil {
		return nil, err
	}

	source, err := NewSource(*sourceURI)
	if err != nil {
		return nil, err
	}

	return &Resolver{
		log,
		cfg,
		source,
	}, nil
}

func (r *Resolver) Event(
	ctx context.Context,
	args *struct {
		On       string
		At       scalar.Offset
		Matching *string
	}) (<-chan *scalar.JSONObject, error) {
	if !acceptTopic(args.On, r.cfg.Topics) {
		return nil, fmt.Errorf("unknown topic: '%s'. Valid topics are: %v", args.On, r.cfg.Topics)
	}
	c := make(chan *scalar.JSONObject)

	ctx = r.log.WithContext(ctx)

	r.source.NewConsumer(ctx, args.On, args.At.Value().Int64()).
		Filter(func(i interface{}) bool {
			return r.acceptMessage(i.(map[string]interface{}), args.Matching)
		}).
		Map(func(i interface{}) interface{} {
			return scalar.NewJSONObject(i.(map[string]interface{}))
		}).
		Subscribe(
			observer.New(
				handlers.NextFunc(func(item interface{}) {
					c <- item.(*scalar.JSONObject)
				}),
				handlers.DoneFunc(func() {
					close(c)
				}),
			))

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

func (r *Resolver) acceptMessage(m map[string]interface{}, predicate *string) bool {
	if predicate == nil {
		return true
	}

	out, err := expr.Eval(*predicate, m)

	if err != nil {
		r.log.
			Warn().
			Object("message", MapAsZerologObject(m)).
			Err(err).
			Msg("⚱️ Failed to filter (message will be dropped)")
		return false
	}

	switch v := out.(type) {
	case bool:
		return v
	default:
		r.log.
			Warn().
			Object("message", MapAsZerologObject(m)).
			Err(fmt.Errorf("incorrect type: %t returned. Expected boolean", out)).
			Msg("⚱️ Failed to filter (message will be dropped)")
		return false
	}
}
