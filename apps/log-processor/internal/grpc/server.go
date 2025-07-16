package grpc

import (
	"context"
	"log-processor/internal/interfaces"
	"log/slog"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	collogspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"

	logcollector "log-processor/internal/controllers/v1/log_collector"
)

type Server struct {
	configService interfaces.ConfigService
	server *grpc.Server
	log *slog.Logger
}

func NewServer(configService interfaces.ConfigService, log *slog.Logger) *Server {
	opts := grpc.ChainUnaryInterceptor(
		// todo: add all these interceptors
		GetRequestLogger(log),
		MapError,
		// grpc.StatsHandler(otelgrpc.NewServerHandler()),
		// grpc.MaxRecvMsgSize(configService.GetConfig().GRPC.MaxReceiveMessageSize),
		// grpc.Creds(insecure.NewCredentials()),
	)
	grpcServer := grpc.NewServer(opts)

	// controllers
	logCollectorControllerV1 := logcollector.New()

	collogspb.RegisterLogsServiceServer(grpcServer, logCollectorControllerV1)

	// todo: do it only for dev
	reflection.Register(grpcServer)

	return &Server{
		configService: configService,
		server: grpcServer,
		log: log,
	}
}

func (s *Server) Start(ctx context.Context) error {
	addr := s.configService.GetConfig().GRPC.Addr
	s.log.Info("Starting gRPC server", "addr", addr)

	lc := net.ListenConfig{}
	listener, err := lc.Listen(ctx, "tcp", addr)
	if err != nil {
		return err
	}

	return s.server.Serve(listener)
}

func (s *Server) Stop() {
	s.server.GracefulStop()
}
