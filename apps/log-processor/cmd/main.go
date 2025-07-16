package main

import (
	"context"
	"sync"

	"io"
	"log/slog"
	_ "net/http/pprof"
	"os"

	"github.com/pkg/errors"

	"log-processor/internal/grpc"
	"log-processor/internal/lib"
	"log-processor/internal/lib/otel"
	"log-processor/internal/services/config"
)

func run(w io.Writer) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := slog.New(slog.NewJSONHandler(w, nil))
	configService := config.NewConfigService()

	if err := configService.LoadConfig(); err != nil {
		return errors.Wrap(err, "could not load config")
	}

	// todo: validate config

	shutdownOtel, err := otel.SetupOTelSDK(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to setup OpenTelemetry SDK")
	}
	defer shutdownOtel(ctx)

	grpcServer := grpc.NewServer(configService, log)

	var shutdownSequenceWg sync.WaitGroup
	shutdownSequenceWg.Add(1)

	return lib.Run(ctx, func() error {
		go func() {
			err := grpcServer.Start(ctx)
			if err != nil {
				// todo: tighten this error handling
				log.ErrorContext(ctx, err.Error())
			}
			shutdownSequenceWg.Done()
		}()

		return nil
	}, func() error {
		cancel()
		shutdownOtel(ctx)
		grpcServer.Stop()

		shutdownSequenceWg.Wait()
		return nil
	})
}

func main() {
	err := run(os.Stdout)
	if err != nil {
		slog.New(slog.NewJSONHandler(os.Stdout, nil)).ErrorContext(context.TODO(), err.Error())
		os.Exit(1)
	}
}
