package server

import "github.com/prometheus/client_golang/prometheus"

var (
	// nolint:gochecknoglobals
	messagesProcessed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "api_messages_processed_total",
			Help: "The total number of processed messages",
		},
		[]string{"stage"},
	)
)

// nolint:gochecknoinits
func init() {
	prometheus.MustRegister(messagesProcessed)
}
