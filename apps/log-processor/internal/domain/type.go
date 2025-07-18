package domain

import v1 "go.opentelemetry.io/proto/otlp/logs/v1"

type LogJob = []*v1.ResourceLogs

type JobResult struct {
	Value string
	MessageHash	[32]byte
}

type AggregationData = map[[32]byte]bool

type Aggregation = map[string]AggregationData
