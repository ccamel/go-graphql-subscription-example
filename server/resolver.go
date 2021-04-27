package server

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/antonmedv/expr"
	"github.com/ccamel/go-graphql-subscription-example/server/log"
	"github.com/ccamel/go-graphql-subscription-example/server/source"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"

	"github.com/ccamel/go-graphql-subscription-example/server/scalar"
)

var (
	ErrUnknownTopic = errors.New("unknown topic")
	ErrUnmarshall   = errors.New("unmarshall error")
)

type Resolver struct {
	log zerolog.Logger
	cfg *Configuration

	source source.Source
}

func NewResolver(cfg *Configuration, log zerolog.Logger) (*Resolver, error) {
	sourceURI, err := url.Parse(cfg.Source)
	if err != nil {
		return nil, fmt.Errorf("url %s failed to be parsed: %w", cfg.Source, err)
	}

	src, err := source.New(sourceURI)
	if err != nil {
		return nil, fmt.Errorf("source %s failed to be created: %w", cfg.Source, err)
	}

	log.
		Info().
		Str("source", src.URI().String()).
		Msgf("Source '%s' configured", src.URI().Scheme)

	return &Resolver{
		log,
		cfg,
		src,
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
		return nil, fmt.Errorf("incorrect topic '%s' (valid topics are: %v): %w", args.On, r.cfg.Topics, ErrUnknownTopic)
	}

	c := make(chan *scalar.JSONObject)

	ctx = r.log.WithContext(ctx)

	r.source.
		NewConsumer(ctx, args.On, args.At.Value().Int64()).
		Map(func(_ context.Context, i interface{}) (interface{}, error) {
			messagesProcessed.
				With(prometheus.Labels{"stage": "received"}).
				Inc()

			return i, nil
		}).
		Filter(func(i interface{}) bool {
			return r.acceptMessage(i.(map[string]interface{}), args.Matching)
		}).
		Map(func(_ context.Context, i interface{}) (interface{}, error) {
			messagesProcessed.
				With(prometheus.Labels{"stage": "accepted"}).
				Inc()

			return i, nil
		}).
		Map(func(_ context.Context, i interface{}) (interface{}, error) {
			return scalar.NewJSONObject(i.(map[string]interface{})), nil
		}).
		Map(func(_ context.Context, i interface{}) (interface{}, error) {
			messagesProcessed.
				With(prometheus.Labels{"stage": "processed"}).
				Inc()

			return i, nil
		}).
		DoOnNext(func(i interface{}) {
			c <- i.(*scalar.JSONObject)
		})

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
			Object("message", log.MapAsZerologObject(m)).
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
			Object("message", log.MapAsZerologObject(m)).
			Err(fmt.Errorf("incorrect type %t returned - expected boolean: %w", out, ErrUnmarshall)).
			Msg("⚱️ Failed to filter (message will be dropped)")

		return false
	}
}
