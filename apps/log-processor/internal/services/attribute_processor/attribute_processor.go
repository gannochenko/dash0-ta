package attribute_processor

import (
	"context"
	"crypto/sha256"
	"errors"
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
	resultCh chan domain.JobResult
	quitCh chan struct{}

	flushTicker *time.Ticker

	aggregation domain.Aggregation
}

func New(config interfaces.ConfigService, log *slog.Logger) *Service {
	return &Service{
		config: config,
		log: log,

		jobsCh: make(chan domain.LogJob, config.GetConfig().JobChannelSize),
		resultCh: make(chan domain.JobResult),

		aggregation: make(domain.Aggregation),
	}
}

func (s *Service) SubmitJob(job domain.LogJob) error {
	select {
	case <-s.quitCh:
		return errors.New("service is shutting down")
	case s.jobsCh <- job:
		return nil
	case <-time.After(5 * time.Second):	
		return errors.New("no workers available")
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

	s.flushTicker = time.NewTicker(time.Duration(s.config.GetConfig().ReportInterval) * time.Second)
}

// Stop gracefully shuts down the worker pool
func (s *Service) Stop() {
	close(s.quitCh)
	close(s.jobsCh)
	close(s.resultCh)
	s.flushTicker.Stop()

	s.workersWg.Wait()
}

// worker processes jobs from the job queue
func (s *Service) worker(ctx context.Context, workerID int) {
	defer s.workersWg.Done()

	s.log.Debug("Starting worker", "worker", workerID)

	for {
		select {
		case <-s.quitCh:
			return
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
		case <-s.quitCh:
			return
		case <-ctx.Done():
			return
		case <-s.flushTicker.C:
			s.printReport()
			s.aggregation = make(domain.Aggregation)
		case value, ok := <-s.resultCh:
			if !ok {
				return
			}

			if _, ok := s.aggregation[value.Value]; !ok {
				s.aggregation[value.Value] = make(domain.AggregationData)
			}
			s.aggregation[value.Value][value.MessageHash] = true
		}
	}
}

func (s *Service) extractAttribute(job domain.LogJob) {
	attributeName := s.config.GetConfig().AttributeName

	for _, resourceLog := range job {
		resourceValue := "[unknown]"

		if resourceLog.Resource != nil {
			for _, attr := range resourceLog.Resource.Attributes {
				if attr.Key == attributeName {
					resourceValue = proto.StringifyValue(attr.Value)
				}
			}
		}

		for _, scopeLog := range resourceLog.ScopeLogs {
			scopeValue := resourceValue

			if scopeLog.Scope != nil {
				for _, attr := range scopeLog.Scope.Attributes {
					if attr.Key == attributeName {
						scopeValue = proto.StringifyValue(attr.Value)
					}
				}
			}

			for _, logRecord := range scopeLog.LogRecords {
				if logRecord.Body == nil {
					// there is no body, skip the log record
					continue
				}

				logValue := scopeValue

				for _, attr := range logRecord.Attributes {
					if attr.Key == attributeName {
						logValue = proto.StringifyValue(attr.Value)
					}
				}

				s.resultCh <- domain.JobResult{
					Value: logValue,
					MessageHash: sha256.Sum256([]byte(proto.StringifyValue(logRecord.Body))),
				}
			}
		}
	}
}

func (s *Service) printReport() {
	fmt.Println("==== Report ====")
	for value, data := range s.aggregation {
		fmt.Printf("Value: %s, Count: %d\n", value, len(data))
	}
}
