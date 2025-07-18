package proto

import (
	"fmt"

	v1 "go.opentelemetry.io/proto/otlp/common/v1"
)

func StringifyValue(value *v1.AnyValue) string {
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
