package domain

type GRPCConfig struct {
	MaxReceiveMessageSize int `envconfig:"MAX_RECEIVE_MESSAGE_SIZE" default:"16777216"`
	Addr                  string `envconfig:"ADDR" default:":443"`
}

type HTTPConfig struct {
	Addr string `envconfig:"ADDR" default:":8080"`
}

type Config struct {
	LogLevel              string `envconfig:"LOG_LEVEL" default:"info" desc:"logging level"`
	AttributeName         string `envconfig:"ATTRIBUTE_NAME" desc:"attribute name"`
	WindowSize            int    `envconfig:"WINDOW_SIZE" default:"1000" desc:"window size in seconds"`
	GRPC                  GRPCConfig `envconfig:"GRPC"`
	HTTP                  HTTPConfig `envconfig:"HTTP"`
	WorkerCount           int    `envconfig:"WORKER_COUNT" default:"5" desc:"number of workers"`
}
