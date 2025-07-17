package attribute_processor

import (
	"context"
	"fmt"
	"log-processor/internal/domain"
	"log-processor/internal/interfaces"
	"log/slog"
	"sync"
	"time"
)

type Service struct {
	config interfaces.ConfigService
	log *slog.Logger

	flushTicker  *time.Ticker
	flushCh      chan struct{}

	workerCount int
	jobsCh        chan domain.AttributeAggregation

	reconcilerTicker *time.Ticker
	reconcileCh chan domain.AttributeAggregation

	workersWg     sync.WaitGroup
	quit        chan struct{}
}

func New(config interfaces.ConfigService, log *slog.Logger) *Service {
	return &Service{
		config: config,
		log: log,

		flushCh: make(chan struct{}),
		reconcileCh: make(chan domain.AttributeAggregation, config.GetConfig().WorkerCount),

		// Initialize worker pool
		workerCount: config.GetConfig().WorkerCount,
		jobsCh:        make(chan domain.AttributeAggregation),
		quit:        make(chan struct{}),
	}
}

func (s *Service) Process(attributes domain.AttributeAggregation) error {
	select {
	case s.jobsCh <- attributes:
		return nil
	case <-s.quit:
		return context.Canceled
	default:
		return context.DeadlineExceeded // can't write to the channel, nobody is reading
	}
}

// Start initializes and starts the worker pool
func (s *Service) Start(ctx context.Context) {
	for i := 0; i < s.workerCount; i++ {
		s.workersWg.Add(1)
		go s.worker(ctx, i)
	}

	s.flushTicker = time.NewTicker(time.Duration(s.config.GetConfig().WindowSize) * time.Second)

    go func() {
        for {
            select {
			case <-ctx.Done():
				return
			case <-s.quit:
				return
            case <-s.flushTicker.C:
                fmt.Println("=== FLUSH SIGNAL ===")
                // Broadcast to all workers
                for i := 0; i < s.workerCount; i++ {
                    select {
                    case s.flushCh <- struct{}{}:
                    default: // Non-blocking send
                    }
                }
            }
        }
    }()

	s.reconcilerTicker = time.NewTicker(100 * time.Millisecond)

	// reconciler
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-s.quit:
				return
			case <-s.reconcilerTicker.C:
				if len(s.reconcileCh) == cap(s.reconcileCh) {
					resultBuffer := make(domain.AttributeAggregation)
					for len(s.reconcileCh) > 0 {
						partialBuffer, ok := <-s.reconcileCh
						if !ok {
							return // Channel closed, reconciler should exit
						}
						s.mergeAggregations(partialBuffer, resultBuffer)
					}

					s.log.Info("Reconciler sends data: \n")

					// for attrKey, values := range resultBuffer {
					// 	s.log.Info("Attribute '%s':", attrKey)
					// 	for value, count := range values {
					// 		log.Printf("'%s': %d occurrences", value, count)
					// 	}
					// }
				}
			}
		}
	}()
}

// Stop gracefully shuts down the worker pool
func (s *Service) Stop() {
	close(s.quit)
	close(s.jobsCh)
	close(s.reconcileCh)
	s.workersWg.Wait()
	s.flushTicker.Stop()
	s.reconcilerTicker.Stop()
}

// worker processes jobs from the job queue
func (s *Service) worker(ctx context.Context, _ int) {
	defer s.workersWg.Done()

	localBuffer := make(domain.AttributeAggregation)

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.quit:
			return
		case _, ok := <-s.flushCh:
			if !ok {
				return // Channel closed, worker should exit
			}
			// log.Printf("Worker %d: Flushing local buffer (%d)", workerID, len(localBuffer))
			select {
			case s.reconcileCh <- localBuffer:
			default:
				// nobody is listening
			}
			localBuffer = make(domain.AttributeAggregation)
		case job, ok := <-s.jobsCh:
			if !ok {
				return // Channel closed, worker should exit
			}
			// log.Printf("Worker %d: Processing job", workerID)
			s.mergeAggregations(job, localBuffer)
		}
	}
}

func (s *Service) mergeAggregations(job domain.AttributeAggregation, localBuffer domain.AttributeAggregation) {
	for key, values := range job {
		for value, count := range values {
			if _, exists := localBuffer[key]; !exists {
				localBuffer[key] = make(map[string]int32)
			}
			localBuffer[key][value] += count
		}
	}
}
