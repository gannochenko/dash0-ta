package attribute_processor

import (
	"context"
	"fmt"
	"log-processor/internal/domain"
	"log-processor/internal/interfaces"
	"log-processor/internal/proto"
	"log/slog"
	"sync"
	"time"
)

type Service struct {
	config interfaces.ConfigService
	log *slog.Logger

	workersWg     sync.WaitGroup

	jobsCh chan domain.LogJob
	valueCh chan string

	flushTicker *time.Ticker

	aggregation map[string]int32
}

func New(config interfaces.ConfigService, log *slog.Logger) *Service {
	return &Service{
		config: config,
		log: log,

		jobsCh: make(chan domain.LogJob),
		valueCh: make(chan string),

		aggregation: make(map[string]int32),
	}
}

func (s *Service) SubmitJob(job domain.LogJob) error {
	select {
	case s.jobsCh <- job:
		return nil
	default:
		return context.DeadlineExceeded // can't write to the channel, nobody is reading
	}
}

// Start initializes and starts the worker pool
func (s *Service) Start(ctx context.Context) {
	s.log.Debug("Starting worker pool", "workerCount", s.config.GetConfig().WorkerCount)

	for i := 0; i < s.config.GetConfig().WorkerCount; i++ {
		s.workersWg.Add(1)
		go s.worker(ctx, i)
	}

	s.workersWg.Add(1)
	go s.aggregator(ctx)

	s.flushTicker = time.NewTicker(time.Duration(s.config.GetConfig().WindowSize) * time.Second)
}

// Stop gracefully shuts down the worker pool
func (s *Service) Stop() {
	close(s.jobsCh)
	close(s.valueCh)
	s.flushTicker.Stop()

	s.workersWg.Wait()
}

// worker processes jobs from the job queue
func (s *Service) worker(ctx context.Context, workerID int) {
	defer s.workersWg.Done()

	s.log.Debug("Starting worker", "worker", workerID)

	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-s.jobsCh:
			if !ok {
				return
			}

			s.log.Debug("Processing job", "job", job, "worker", workerID)

			s.extractAttribute(job)
		}
	}
}

func (s *Service) aggregator(ctx context.Context) {
	defer s.workersWg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.flushTicker.C:
			s.printAggregation()
			s.aggregation = make(map[string]int32)
		case value, ok := <-s.valueCh:
			if !ok {
				return
			}

			s.aggregation[value]++
		}
	}
}

func (s *Service) extractAttribute(job domain.LogJob) {
	attributeName := s.config.GetConfig().AttributeName

	for _, resourceLog := range job {
		if resourceLog.Resource != nil {
			for _, attr := range resourceLog.Resource.Attributes {
				if attr.Key == attributeName {
					s.valueCh <- proto.StringifyAttributeValue(attr.Value)
				}
			}
		}

		for _, scopeLog := range resourceLog.ScopeLogs {
			if scopeLog.Scope != nil {
				for _, attr := range scopeLog.Scope.Attributes {
					if attr.Key == attributeName {
						s.valueCh <- proto.StringifyAttributeValue(attr.Value)
					}
				}
			}

			for _, logRecord := range scopeLog.LogRecords {
				for _, attr := range logRecord.Attributes {
					if attr.Key == attributeName {
						s.valueCh <- proto.StringifyAttributeValue(attr.Value)
					}
				}
			}
		}
	}
}

func (s *Service) printAggregation() {
	fmt.Println("==== Aggregation ====")
	for value, count := range s.aggregation {
		fmt.Printf("Value: %s, Count: %d\n", value, count)
	}
}
