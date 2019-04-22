package server

import (
	"context"
	"encoding/json"
	"io"

	"github.com/reactivex/rxgo"

	"github.com/rs/zerolog"

	uuid "github.com/satori/go.uuid"
	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	consumerID uuid.UUID
	brokers    []string
	topic      string
	offset     int64
	ctx        context.Context
	log        zerolog.Logger
}

func NewConsumer(
	ctx context.Context,
	brokers []string,
	topic string,
	offset int64) *Consumer {
	id := uuid.NewV4()

	return &Consumer{
		id,
		brokers,
		topic,
		offset,
		ctx,
		zerolog.
			Ctx(ctx).
			With().
			Str("consumerID", id.String()).
			Logger(),
	}
}

func (c Consumer) AsObservable() rxgo.Observable {
	return rxgo.Create(func(emitter rxgo.Observer, disposed bool) {
		defer func() {
			emitter.OnDone()

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

			emitter.OnError(err)

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

			c.log.
				Info().
				Object("message", KafkaMessageAsZerologObject(m)).
				Msg("↩️ Sending message to subscriber")

			emitter.OnNext(v)
		}
	})
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
