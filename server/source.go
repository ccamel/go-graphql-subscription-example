package server

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/reactivex/rxgo/v2"
)

// Source specifies types which are able to provide a source of events through an Observable.
type Source interface {
	URI() *url.URL
	// NewConsumer returns a new observable consuming messages from the this source, from a topic, starting
	// at provided offset (if supported).
	NewConsumer(ctx context.Context, topic string, offset int64) rxgo.Observable
}

type SourceFactory func(uri *url.URL) (Source, error)

// nolint:gochecknoglobals
var sourceFactories = make(map[string]SourceFactory)

// RegisterSourceFactory registers a new source factory for the considered scheme.
func RegisterSourceFactory(scheme string, factory SourceFactory) {
	sourceFactories[scheme] = factory
}

// NewSource returns a new instance of source given the uri.
// The uri contains all the required information to perform a connection to the source endpoint.
func NewSource(uri *url.URL) (Source, error) {
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

	return nil, fmt.Errorf("scheme %s is not supported. Available are: ", strings.Join(keys, ","))
}
