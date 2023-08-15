package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	nsq "github.com/nsqio/go-nsq"
	rxgo "github.com/reactivex/rxgo/v2"
	"github.com/rs/zerolog"
	uuid "github.com/satori/go.uuid"

	"github.com/ccamel/go-graphql-subscription-example/server/log"
	"github.com/ccamel/go-graphql-subscription-example/server/source"
)

type nsqSource struct {
	uri         *url.URL
	config      *nsq.Config
	lookupdAddr string
}

func (s nsqSource) URI() *url.URL {
	return s.uri
}

func (s nsqSource) NewConsumer(ctx context.Context, topic string, offset int64) rxgo.Observable {
	id := uuid.NewV4()

	c := nsqConsumer{
		id,
		topic,
		ctx,
		zerolog.
			Ctx(ctx).
			With().
			Str("consumerID", id.String()).
			Logger(),
	}

	return rxgo.Create([]rxgo.Producer{func(ctx context.Context, next chan<- rxgo.Item) {
		defer func() {
			c.log.
				Info().
				Msg("⛔ Consumer stopped")
		}()

		c.log.
			Info().
			Str("topic", c.topic).
			Int64("offset", offset).
			Msg("▶️ Consumer started")

		q, _ := nsq.NewConsumer(topic, fmt.Sprintf("ch-%s", c.consumerID.String()), s.config)
		q.AddHandler(nsq.HandlerFunc(func(m *nsq.Message) error {
			v, success := unmarshalNsqMessage(c, m)
			if !success {
				return nil
			}

			c.log.
				Info().
				Object("message", NsqMessageAsZerologObject(m)).
				Msg("↩️ Sending message to subscriber")

			next <- rxgo.Of(v)

			return nil
		}))
		err := q.ConnectToNSQLookupd(s.lookupdAddr)
		if err != nil {
			c.log.
				Warn().
				Err(err).
				Msg("Error when connecting to nsq server")

			next <- rxgo.Error(err)

			return
		}

		go func() {
			<-ctx.Done()
			q.Stop()
		}()

		<-q.StopChan
	}}, rxgo.WithContext(c.ctx)) //nolint:contextcheck
}

// NsqMessageAsZerologObject converts a NSQ message into a LogObjectMarshaler.
func NsqMessageAsZerologObject(message *nsq.Message) log.LoggerFunc {
	return func(e *zerolog.Event) {
		e.
			Bytes("id", message.ID[:]).
			Int64("timestamp", message.Timestamp).
			Uint16("attempts", message.Attempts).
			Int("size", len(message.Body))
	}
}

type nsqConsumer struct {
	consumerID uuid.UUID
	topic      string
	ctx        context.Context
	log        zerolog.Logger
}

func newNsqSource(uri *url.URL) (source.Source, error) {
	config, err := makeNsqOptions(uri)
	if err != nil {
		return nil, err
	}

	lookupdAddr := uri.Host
	if lookupdAddr == "" {
		lookupdAddr = "localhost:4161"
	}

	return &nsqSource{
		uri:         uri,
		config:      config,
		lookupdAddr: lookupdAddr,
	}, nil
}

//nolint:unparam
func makeNsqOptions(source *url.URL) (*nsq.Config, error) {
	config := nsq.NewConfig()

	clientID := source.Query().Get("client_id")
	if clientID != "" {
		_ = config.Set("client_id", clientID)
	}

	return config, nil
}

// unmarshalNsqMessage parses the given message read from NSQ into a JSON map.
func unmarshalNsqMessage(c nsqConsumer, m *nsq.Message) (map[string]interface{}, bool) {
	var v map[string]interface{}

	if err := json.Unmarshal(m.Body, &v); err != nil {
		c.log.
			Warn().
			Object("message", NsqMessageAsZerologObject(m)).
			Err(err).
			Msg("⚱️ Failed to unmarshal message (message will be dropped)")

		return nil, false
	}

	return v, true
}

//nolint:gochecknoinits
func init() {
	source.RegisterFactory("nsq", newNsqSource)
}
