package server

import (
	"context"
	"net/url"
	"time"

	"github.com/go-redis/redis"
	"github.com/reactivex/rxgo/observable"
	"github.com/reactivex/rxgo/observer"
	"github.com/robinjoseph08/redisqueue"
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

func newRedisSource(uri *url.URL) (Source, error) {
	opt, err := makeRedisOptions(uri)

	if err != nil {
		return nil, err
	}

	return &redisSource{
		uri:             uri,
		consumerOptions: opt,
	}, nil
}

func (s redisSource) NewConsumer(ctx context.Context, topic string, offset int64) observable.Observable {
	id := uuid.NewV4()

	l := zerolog.
		Ctx(ctx).
		With().
		Str("consumerID", id.String()).
		Logger()

	c, errored := redisqueue.NewConsumerWithOptions(s.consumerOptions)

	return observable.Create(func(emitter *observer.Observer, disposed bool) {
		defer func() {
			emitter.OnDone()

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

			emitter.OnError(errored)

			return
		}

		c.Register(topic, func(message *redisqueue.Message) error {
			v, success := unmarshalRedisMessage(message)
			if success {
				l.
					Info().
					Dict("message", zerolog.Dict().Fields(v)).
					Msg("↩️ Sending message to subscriber")

				emitter.OnNext(v)
			}

			return nil
		})

		go func() {
			<-ctx.Done()
			c.Shutdown()
		}()

		c.Run()
	})
}

// nolint:unparam
func makeRedisOptions(source *url.URL) (*redisqueue.ConsumerOptions, error) {
	options := &redisqueue.ConsumerOptions{
		VisibilityTimeout: 60 * time.Second,
		BlockingTimeout:   5 * time.Second,
		ReclaimInterval:   1 * time.Second,
		BufferSize:        100,
		Concurrency:       10,
		RedisOptions: &redis.Options{
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
	RegisterSourceFactory("redis", newRedisSource)
}
