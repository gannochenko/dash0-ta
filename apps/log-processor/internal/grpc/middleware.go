package grpc

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func MapError(
	ctx context.Context,
	req interface{},
	_ *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	resp, err := handler(ctx, req)

	if err != nil {
		code := status.Code(err)

		if code == codes.Unknown {
			return nil, status.Error(codes.Internal, err.Error())
		}

		return nil, err
	}

	return resp, nil
}

func GetRequestLogger(log *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		resp, err := handler(ctx, req)

		b := bytes.NewBuffer([]byte{})
		_ = json.NewEncoder(b).Encode(req)

		requestBody := b.String()

		grpcRequestAttr := slog.String("grpc_request", info.FullMethod)
		requestDataAttr := slog.String("request_data", strings.Trim(requestBody, " \r\n"))

		if err != nil {
			code := status.Code(err)
			switch code {
			case codes.InvalidArgument, codes.NotFound, codes.AlreadyExists, codes.PermissionDenied:
				log.WarnContext(ctx, errors.Wrap(err, "request handled with error").Error(), grpcRequestAttr, requestDataAttr)
			default:
				log.ErrorContext(ctx, errors.Wrap(err, "request handled with error").Error(), grpcRequestAttr, requestDataAttr)
			}
		} else {
			log.DebugContext(ctx, "request handled", grpcRequestAttr, requestDataAttr)
		}

		return resp, err
	}
}
