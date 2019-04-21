package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/antonmedv/expr"

	"github.com/rs/zerolog"

	"github.com/ccamel/go-graphql-subscription-example/server/scalar"

	uuid "github.com/satori/go.uuid"
	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	consumerID uuid.UUID
	brokers    []string
	topic      string
	offset     int64
	predicate  *string
	ctx        context.Context
	channel    chan<- *scalar.JSONObject
	log        zerolog.Logger
}

func NewConsumer(
	ctx context.Context,
	brokers []string,
	topic string,
	offset int64,
	matching *string,
	channel chan<- *scalar.JSONObject) (*Consumer, error) {
	id := uuid.NewV4()

	return &Consumer{
		id,
		brokers,
		topic,
		offset,
		matching,
		ctx,
		channel,
		zerolog.
			Ctx(ctx).
			With().
			Str("consumerID", id.String()).
			Logger(),
	}, nil
}

func (c Consumer) Start() {
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
		return
	}

	defer func() {
		_ = r.Close()
	}()

	for {
		m, err := r.ReadMessage(c.ctx)
		if err != nil {
			switch err {
			case io.EOF:
			default:
				c.log.
					Warn().
					Err(err).
					Msg("❌ Error when reading message")
			}
			break
		}

		v, success := c.unmarshal(m)
		if !success {
			continue
		}

		if !c.matches(m, v) {
			continue
		}

		c.log.
			Info().
			Object("message", KafkaMessageAsZerologObject(m)).
			Msg("↩️ Sending message to subscriber")

		c.channel <- scalar.NewJSONObject(v)
	}

	c.log.
		Info().
		Msg("⛔ Consumer stopped")
}

// unmarshal parses the given kafka message into a JSON map.
func (c Consumer) unmarshal(m kafka.Message) (map[string]interface{}, bool) {
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

func (c Consumer) matches(m kafka.Message, v map[string]interface{}) bool {
	if c.predicate == nil {
		return true
	}

	out, err := expr.Eval(*c.predicate, v)

	if err != nil {
		c.log.
			Warn().
			Object("message", KafkaMessageAsZerologObject(m)).
			Err(err).
			Msg("⚱️ Failed to filter (message will be dropped)")
		return false
	}

	switch v := out.(type) {
	case bool:
		return v
	default:
		c.log.
			Warn().
			Object("message", KafkaMessageAsZerologObject(m)).
			Err(fmt.Errorf("incorrect type: %t returned. Expected boolean", out)).
			Msg("⚱️ Failed to filter (message will be dropped)")
		return false
	}

}
