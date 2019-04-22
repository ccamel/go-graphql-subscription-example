package server

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/segmentio/kafka-go"
)

// LoggerFunc turns a function into an a zerolog marshaller.
type LoggerFunc func(e *zerolog.Event)

// MarshalZerologObject makes the LoggerFunc type a LogObjectMarshaler.
func (f LoggerFunc) MarshalZerologObject(e *zerolog.Event) {
	f(e)
}

// AsEventTraitZerologObject converts a kafka message into a LogObjectMarshaler.
func KafkaMessageAsZerologObject(message kafka.Message) LoggerFunc {
	return LoggerFunc(func(e *zerolog.Event) {
		e.
			Str("topic", message.Topic).
			Int64("offset", message.Offset).
			Time("time", message.Time).
			Int("size", len(message.Value))
	})
}

// MapAsZerologObject converts a map into a LogObjectMarshaler.
func MapAsZerologObject(m map[string]interface{}) LoggerFunc {
	return LoggerFunc(func(e *zerolog.Event) {
		e.
			Fields(m)
	})
}

func NewLogger() zerolog.Logger {
	return zerolog.New(os.Stderr).With().Timestamp().Logger()
}
