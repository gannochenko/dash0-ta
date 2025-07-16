package logcollector

import (
	"context"
	"fmt"
	"log"

	collogspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	v1 "go.opentelemetry.io/proto/otlp/common/v1"
)

type Controller struct {
	collogspb.UnimplementedLogsServiceServer
}

func New() collogspb.LogsServiceServer {
	return &Controller{}
}

func (l *Controller) Export(ctx context.Context, request *collogspb.ExportLogsServiceRequest) (*collogspb.ExportLogsServiceResponse, error) {
	// Map structure: attribute_key -> value_string -> count
	attributes := make(map[string]map[string]int32)

	recordAttribute := func(key, value string) {
		if attributes[key] == nil {
			attributes[key] = make(map[string]int32)
		}
		attributes[key][value]++
	}
	
	for _, resourceLog := range request.ResourceLogs {
		if resourceLog.Resource != nil {
			for _, attr := range resourceLog.Resource.Attributes {
				value := l.extractValue(attr.Value)
				recordAttribute(attr.Key, value)
			}
		}

		for _, scopeLog := range resourceLog.ScopeLogs {
			if scopeLog.Scope != nil {
				for _, attr := range scopeLog.Scope.Attributes {
					value := l.extractValue(attr.Value)
					recordAttribute(attr.Key, value)
				}
			}
			
			for _, logRecord := range scopeLog.LogRecords {
				for _, attr := range logRecord.Attributes {
					value := l.extractValue(attr.Value)
					recordAttribute(attr.Key, value)
				}
			}
		}
	}

	// Log the final attribute statistics
	for attrKey, values := range attributes {
		log.Printf("Attribute '%s':", attrKey)
		for value, count := range values {
			log.Printf("  '%s': %d occurrences", value, count)
		}
	}

	return &collogspb.ExportLogsServiceResponse{}, nil
}

func (l *Controller) extractValue(value *v1.AnyValue) string {
	if value == nil {
		return ""
	}

	switch v := value.Value.(type) {
	case *v1.AnyValue_StringValue:
		return v.StringValue
	case *v1.AnyValue_BoolValue:
		if v.BoolValue {
			return "true"
		}
		return "false"
	case *v1.AnyValue_IntValue:
		return fmt.Sprintf("%d", v.IntValue)
	case *v1.AnyValue_DoubleValue:
		return fmt.Sprintf("%f", v.DoubleValue)
	case *v1.AnyValue_ArrayValue:
		return "[array]" // todo: support this
	case *v1.AnyValue_KvlistValue:
		return "[kvlist]" // todo: support this
	case *v1.AnyValue_BytesValue:
		return string(v.BytesValue)
	default:
		return "[unknown]"
	}
}