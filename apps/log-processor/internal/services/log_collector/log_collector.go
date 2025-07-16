package log_collector

import (
	collogspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
)

type LogCollector struct {
	collogspb.UnimplementedLogsServiceServer
}

func NewLogCollector() *LogCollector {
	return &LogCollector{}
}
