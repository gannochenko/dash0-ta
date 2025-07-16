package logcollector

import (
	"context"

	collogspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
)

type Controller struct {
	collogspb.UnimplementedLogsServiceServer
}

func New() collogspb.LogsServiceServer {
	return &Controller{}
}

func (l *Controller) Export(ctx context.Context, request *collogspb.ExportLogsServiceRequest) (*collogspb.ExportLogsServiceResponse, error) {
	// todo: implement

	return &collogspb.ExportLogsServiceResponse{}, nil
}
