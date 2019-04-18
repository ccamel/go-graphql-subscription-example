package server

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func StartCommand() {
	cfg := &Configuration{}

	rootCmd := &cobra.Command{
		Use:   "go-graphql-subscription-example",
		Short: "Service exposing a graphQL endpoint for subscribing to kafka topics",
		Long: `Service that exposes a graphQL endpoint allowing client to subscribe 
to kafka topics.
		
					See https://github.com/ccamel/go-graphql-subscription-example`,
		Run: func(cmd *cobra.Command, args []string) {
			StartServer(cfg)
		},
	}

	rootCmd.Flags().Uint16Var(&cfg.Port, "port", 8000, "The listening port")
	rootCmd.Flags().StringSliceVar(&cfg.Brokers, "brokers", []string{"localhost:9092"}, "The list of broker addresses used to connect to the kafka cluster")
	rootCmd.Flags().StringSliceVar(&cfg.Topics, "topics", []string{"foo"}, "The list of kafka topics that subscribers can consume")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
