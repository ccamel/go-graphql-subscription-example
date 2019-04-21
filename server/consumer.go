package server

import (
	"context"
	"encoding/json"
	"io"

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
	ctx        context.Context
	channel    chan<- *scalar.JSONObject
}

func NewConsumer(
	ctx context.Context,
	brokers []string,
	topic string,
	offset int64,
	channel chan<- *scalar.JSONObject) *Consumer {
	return &Consumer{
		uuid.NewV4(),
		brokers,
		topic,
		offset,
		ctx,
		channel,
	}
}

func (c Consumer) Start() {
	log :=
		zerolog.
			Ctx(c.ctx).
			With().
			Str("consumerID", c.consumerID.String()).
			Logger()

	log.
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

	err := r.SetOffset(c.offset)
	if err != nil {
		log.
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
				log.
					Warn().
					Err(err).
					Msg("❌ Error when reading message")
			}
			break
		}

		// parse to json
		var v map[string]interface{}
		if err := json.Unmarshal(m.Value, &v); err != nil {
			log.
				Warn().
				Object("message", KafkaMessageAsZerologObject(m)).
				Err(err).
				Msg("⚱️ Failed to unmarshal message (message will be dropped)")

			continue
		}

		log.
			Info().
			Object("message", KafkaMessageAsZerologObject(m)).
			Msg("↩️ Sending message to subscriber")

		c.channel <- scalar.NewJSONObject(v)
	}

	log.
		Info().
		Msg("⛔ Consumer stopped")
}
