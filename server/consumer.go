package server

import (
	"context"
	"encoding/json"
	"io"

	graphql "github.com/ccamel/go-graphql-subscription-example/server/scalar"
	"github.com/segmentio/kafka-go"
)

func consume(ctx context.Context, channel chan<- *graphql.JSONObject) {
	log.
		Info().
		Msg("Start consuming messages")

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{"localhost:9092"},
		Topic:     "in",
		Partition: 0,
		MinBytes:  10e3, // 10KB
		MaxBytes:  10e6, // 10MB
	})

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

		channel <- graphql.New(v)
	}
}
