package server

import (
	"github.com/spf13/cobra"
)

func StartCommand() {
	l := NewLogger()
	cfg := &Configuration{}

	rootCmd := &cobra.Command{
		Use:   "go-graphql-subscription-example",
		Short: "Service exposing a graphQL endpoint for subscribing to different kind of stream sources.",
		Long: `Service that exposes a graphQL endpoint allowing client to subscribe to different kind of stream sources.
Sources currently supported are:
 - Kafka: The distributed streaming platform.
		
 See https://github.com/ccamel/go-graphql-subscription-example`,
		Run: func(cmd *cobra.Command, args []string) {
			server := NewServer(cfg)

			server.Start()
		},
	}

	f := rootCmd.Flags()
	f.Uint16Var(&cfg.Port, "port", 8000, "The listening port")
	f.StringSliceVar(&cfg.Topics, "topics", []string{"foo"}, "The list of topics/stream names that subscribers can consume")
	f.StringVar(&cfg.Source, "source", "kafka:?brokers=localhost:9092", "The URI of the source to connect to")

	if err := rootCmd.Execute(); err != nil {
		l.
			Fatal().
			Err(err).
			Msg("Execution failed")
	}
}
