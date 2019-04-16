package server

type Configuration struct {
	// The port the server will listen to.
	Port uint16
	// The list of broker addresses used to connect to the kafka cluster.
	Brokers []string
	// The list of topics subscribers can consume.
	Topics []string
}
