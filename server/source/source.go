package source

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/reactivex/rxgo/v2"
)

var (
	ErrIncorrectScheme = errors.New("incorrect scheme")
)

// Source specifies types which are able to provide a source of events through an Observable.
type Source interface {
	URI() *url.URL
	// NewConsumer returns a new observable consuming messages from the this source, from a topic, starting
	// at provided offset (if supported).
	NewConsumer(ctx context.Context, topic string, offset int64) rxgo.Observable
}

// Factory specifies functions able to create sources from an URI.
type Factory func(uri *url.URL) (Source, error)

// sourceFactories constains all the registered source factories.
// nolint:gochecknoglobals
var sourceFactories = make(map[string]Factory)

// RegisterFactory registers a new source factory for the considered scheme.
func RegisterFactory(scheme string, factory Factory) {
	sourceFactories[scheme] = factory
}

// New returns a new instance of source given the uri.
// The uri contains all the required information to perform a connection to the source endpoint.
func New(uri *url.URL) (Source, error) {
	for scheme, factory := range sourceFactories {
		if uri.Scheme == scheme {
			return factory(uri)
		}
	}

	keys := make([]string, len(sourceFactories))

	i := 0

	for k := range sourceFactories {
		keys[i] = k
		i++
	}

	return nil, fmt.Errorf("scheme %s is not supported (available are: %s): %w",
		uri.Scheme, strings.Join(keys, ","), ErrIncorrectScheme)
}
