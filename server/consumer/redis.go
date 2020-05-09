package consumer

import (
	"context"
	"net/url"
	"time"

	"github.com/ccamel/go-graphql-subscription-example/server/source"
	"github.com/reactivex/rxgo/v2"
	"github.com/robinjoseph08/redisqueue/v2"
	"github.com/rs/zerolog"

	uuid "github.com/satori/go.uuid"
)

type redisSource struct {
	uri             *url.URL
	consumerOptions *redisqueue.ConsumerOptions
}

func (s redisSource) URI() *url.URL {
	return s.uri
}

func newRedisSource(uri *url.URL) (source.Source, error) {
	opt, err := makeRedisOptions(uri)

	if err != nil {
		return nil, err
	}

	return &redisSource{
		uri:             uri,
		consumerOptions: opt,
	}, nil
}

func (s redisSource) NewConsumer(ctx context.Context, topic string, offset int64) rxgo.Observable {
	id := uuid.NewV4()

	l := zerolog.
		Ctx(ctx).
		With().
		Str("consumerID", id.String()).
		Logger()

	c, errored := redisqueue.NewConsumerWithOptions(s.consumerOptions)

	return rxgo.Create([]rxgo.Producer{func(ctx context.Context, next chan<- rxgo.Item) {
		defer func() {
			l.
				Info().
				Msg("⛔ Consumer stopped")
		}()

		l.
			Info().
			Str("topic", topic).
			Int64("offset", offset).
			Msg("▶️ Consumer started")

		if errored != nil {
			l.
				Warn().
				Err(errored).
				Msg("Error when connecting to redis server")

			next <- rxgo.Error(errored)

			return
		}

		c.Register(topic, func(message *redisqueue.Message) error {
			v, success := unmarshalRedisMessage(message)
			if success {
				l.
					Info().
					Dict("message", zerolog.Dict().Fields(v)).
					Msg("↩️ Sending message to subscriber")

				next <- rxgo.Of(v)
			}

			return nil
		})

		go func() {
			<-ctx.Done()
			c.Shutdown()
		}()

		c.Run()
	}}, rxgo.WithContext(ctx))
}

// nolint:unparam
func makeRedisOptions(source *url.URL) (*redisqueue.ConsumerOptions, error) {
	options := &redisqueue.ConsumerOptions{
		VisibilityTimeout: 60 * time.Second,
		BlockingTimeout:   5 * time.Second,
		ReclaimInterval:   1 * time.Second,
		BufferSize:        100,
		Concurrency:       10,
		RedisOptions: &redisqueue.RedisOptions{
			Addr: ":6379",
		},
	}

	if source.Host != "" {
		options.RedisOptions.Addr = source.Host
	}

	name := source.Query().Get("name")
	if name != "" {
		options.Name = name
	}

	return options, nil
}

// unmarshalRedisMessage parses the given message read from Redis into a JSON map.
func unmarshalRedisMessage(m *redisqueue.Message) (map[string]interface{}, bool) {
	return m.Values, true
}

// nolint:gochecknoinits
func init() {
	source.RegisterSourceFactory("redis", newRedisSource)
}
