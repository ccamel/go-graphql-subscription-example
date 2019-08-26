package server

type Configuration struct {
	// The port the server will listen to.
	Port uint16
	// The list of topics/stream names (Kafka/Redis) subscribers can consume.
	Topics []string
	// The server URI used to connect to the stream source (either Kafka or Redis).
	Source string
}
