package logcollector

import (
	"context"
	"log-processor/internal/interfaces"

	"github.com/pkg/errors"
	collogspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
)

type Controller struct {
	collogspb.UnimplementedLogsServiceServer
	attributeProcessorService interfaces.AttributeProcessorService
}

func New(attributeProcessorService interfaces.AttributeProcessorService) collogspb.LogsServiceServer {
	return &Controller{
		attributeProcessorService: attributeProcessorService,
	}
}

func (l *Controller) Export(ctx context.Context, request *collogspb.ExportLogsServiceRequest) (*collogspb.ExportLogsServiceResponse, error) {
	err := l.attributeProcessorService.SubmitJob(request.ResourceLogs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to send job for processing")
	}

	return &collogspb.ExportLogsServiceResponse{}, nil
}
