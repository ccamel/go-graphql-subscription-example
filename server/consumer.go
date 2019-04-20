package server

import (
	"context"
	"encoding/json"
	"io"

	"github.com/rs/zerolog"

	graphql "github.com/ccamel/go-graphql-subscription-example/server/scalar"
	"github.com/segmentio/kafka-go"
)

func consume(ctx context.Context, channel chan<- *graphql.JSONObject) {
	topic := ctx.Value(topicKey).(string)
	offset := ctx.Value(offsetKey).(graphql.Offset).Value()

	log :=
		zerolog.
			Ctx(ctx).
			With().
			Str("subscriptionID", ctx.Value(subscriptionID).(string)).
			Logger()

	log.
		Info().
		Str("topic", topic).
		Str("offset", offset.String()).
		Msg("Start consuming messages")

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   ctx.Value(brokersKey).([]string),
		Topic:     topic,
		Partition: 0,

		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	err := r.SetOffset(offset.Int64())
	if err != nil {
		log.
			Warn().
			Err(err).
			Str("offset", offset.String()).
			Msg("Error when setting offset")
		return
	}

	defer func() {
		_ = r.Close()
	}()

	for {
		m, err := r.ReadMessage(ctx)
		if err != nil {
			switch err {
			case io.EOF:
				log.
					Info().
					Msg("Stop consuming")
			default:
				log.
					Warn().
					Err(err).
					Msg("Error when reading message")
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
				Msg("Failed to unmarshal message")

			continue
		}

		log.
			Info().
			Object("message", KafkaMessageAsZerologObject(m)).
			Msg("Sending new message to subscriber")

		channel <- graphql.NewJSONObject(v)
	}
}
