package server

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/segmentio/kafka-go"
)

var log zerolog.Logger

// LoggerFunc turns a function into an a zerolog marshaller.
type LoggerFunc func(e *zerolog.Event)

// MarshalZerologObject makes the LoggerFunc type a LogObjectMarshaler.
func (f LoggerFunc) MarshalZerologObject(e *zerolog.Event) {
	f(e)
}

// AsEventTraitZerologObject converts a
func KafkaMessageAsZerologObject(message kafka.Message) LoggerFunc {
	return LoggerFunc(func(e *zerolog.Event) {
		e.
			Str("topic", message.Topic).
			Int64("offset", message.Offset).
			Time("time", message.Time).
			Int("size", len(message.Value))
	})
}

func init() {
	log = zerolog.New(os.Stderr).With().Timestamp().Logger()
}
