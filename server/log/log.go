package log

import (
	"os"

	"github.com/rs/zerolog"
)

// LoggerFunc turns a function into an a zerolog marshaller.
type LoggerFunc func(e *zerolog.Event)

// MarshalZerologObject makes the LoggerFunc type a LogObjectMarshaler.
func (f LoggerFunc) MarshalZerologObject(e *zerolog.Event) {
	f(e)
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
