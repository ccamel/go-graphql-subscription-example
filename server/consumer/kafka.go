package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"

	rxgo "github.com/reactivex/rxgo/v2"
	"github.com/rs/zerolog"
	uuid "github.com/satori/go.uuid"
	kafka "github.com/segmentio/kafka-go"

	"github.com/ccamel/go-graphql-subscription-example/server/log"
	"github.com/ccamel/go-graphql-subscription-example/server/source"
)

var ErrNoBrokerSpecified = errors.New("no broker specified")

type kafkaSource struct {
	uri     *url.URL
	brokers []string
}

func (s kafkaSource) URI() *url.URL {
	return s.uri
}

type kafkaConsumer struct {
	consumerID uuid.UUID
	brokers    []string
	topic      string
	offset     int64
	ctx        context.Context
	log        zerolog.Logger
}

func newKafkaSource(uri *url.URL) (source.Source, error) {
	brokers, err := parseKafkaBrokers(uri)
	if err != nil {
		return nil, err
	}

	return &kafkaSource{
		uri:     uri,
		brokers: brokers,
	}, nil
}

func (s kafkaSource) NewConsumer(ctx context.Context, topic string, offset int64) rxgo.Observable {
	id := uuid.NewV4()

	c := kafkaConsumer{
		id,
		s.brokers,
		topic,
		offset,
		ctx,
		zerolog.
			Ctx(ctx).
			With().
			Str("consumerID", id.String()).
			Logger(),
	}

	return makeObservableFromKafkaConsumer(c)
}

// KafkaMessageAsZerologObject converts a kafka message into a LogObjectMarshaler.
func KafkaMessageAsZerologObject(message kafka.Message) log.LoggerFunc {
	return func(e *zerolog.Event) {
		e.
			Str("topic", message.Topic).
			Int64("offset", message.Offset).
			Time("time", message.Time).
			Int("size", len(message.Value))
	}
}

func parseKafkaBrokers(source *url.URL) ([]string, error) {
	brokersStr := source.Query().Get("brokers")

	if brokersStr == "" {
		return nil, fmt.Errorf("incorrect source URL %s: %w", source.String(), ErrNoBrokerSpecified)
	}

	return strings.Split(brokersStr, ","), nil
}

// unmarshalKafkaMessage parses the given kafka message into a JSON map.
func unmarshalKafkaMessage(c kafkaConsumer, m kafka.Message) (map[string]interface{}, bool) {
	var v map[string]interface{}

	if err := json.Unmarshal(m.Value, &v); err != nil {
		c.log.
			Warn().
			Object("message", KafkaMessageAsZerologObject(m)).
			Err(err).
			Msg("⚱️ Failed to unmarshal message (message will be dropped)")

		return nil, false
	}

	return v, true
}

func makeObservableFromKafkaConsumer(c kafkaConsumer) rxgo.Observable {
	return rxgo.Create([]rxgo.Producer{func(_ context.Context, next chan<- rxgo.Item) {
		defer func() {
			c.log.
				Info().
				Msg("⛔ Consumer stopped")
		}()

		c.log.
			Info().
			Str("topic", c.topic).
			Int64("offset", c.offset).
			Msg("▶️ Consumer started")

		r := kafka.NewReader(kafka.ReaderConfig{
			Brokers:   c.brokers,
			Topic:     c.topic,
			Partition: 0,

			MinBytes: 10e3, // 10KB
			MaxBytes: 10e6, // 10MB
		})

		if err := r.SetOffset(c.offset); err != nil {
			c.log.
				Warn().
				Err(err).
				Int64("offset", c.offset).
				Msg("Error when setting offset")

			next <- rxgo.Error(err)

			return
		}

		defer func() {
			_ = r.Close()
		}()

		for {
			m, err := r.ReadMessage(c.ctx) //nolint:contextcheck
			if err != nil {
				if !errors.Is(err, io.EOF) {
					c.log.
						Warn().
						Err(err).
						Msg("❌ Error when reading message")
				}

				break
			}

			v, success := unmarshalKafkaMessage(c, m)
			if !success {
				continue
			}

			c.log.
				Info().
				Object("message", KafkaMessageAsZerologObject(m)).
				Msg("↩️ Sending message to subscriber")

			next <- rxgo.Of(v)
		}
	}}, rxgo.WithContext(c.ctx))
}

//nolint:gochecknoinits
func init() {
	source.RegisterFactory("kafka", newKafkaSource)
}
