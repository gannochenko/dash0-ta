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
	ReportInterval        int    `envconfig:"REPORT_INTERVAL" default:"10" desc:"report interval in seconds"`
	GRPC                  GRPCConfig `envconfig:"GRPC"`
	HTTP                  HTTPConfig `envconfig:"HTTP"`
	WorkerCount           int    `envconfig:"WORKER_COUNT" default:"5" desc:"number of workers"`
	JobChannelSize        int    `envconfig:"JOB_CHANNEL_SIZE" default:"1000" desc:"size of the job channel"`
}
