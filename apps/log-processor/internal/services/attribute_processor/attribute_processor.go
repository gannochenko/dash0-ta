package attribute_processor

import (
	"context"
	"fmt"
	"log"
	"log-processor/internal/domain"
	"log-processor/internal/interfaces"
	"sync"
	"time"
)

type Service struct {
	config interfaces.ConfigService

	flushTicker  *time.Ticker
	flushCh      chan struct{}

	workerCount int
	jobsCh        chan domain.AttributeAggregation
	reconcileCh chan domain.AttributeAggregation

	workersWg     sync.WaitGroup
	quit        chan struct{}
}

func New(config interfaces.ConfigService) *Service {
	return &Service{
		config: config,

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
            case <-s.flushTicker.C:
                fmt.Println("=== FLUSH SIGNAL ===")
                // Broadcast to all workers
                for i := 0; i < s.workerCount; i++ {
                    select {
                    case s.flushCh <- struct{}{}:
                    default: // Non-blocking send
                    }
                }
			case <-s.quit:
				return
            }
        }
    }()

	// reconciler
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-s.quit:
				return
			default:
				if len(s.reconcileCh) == cap(s.reconcileCh) {
					resultBuffer := make(domain.AttributeAggregation)
					for len(s.reconcileCh) > 0 {
						partialBuffer, ok := <-s.reconcileCh
						if !ok {
							return // Channel closed, reconciler should exit
						}
						s.mergeAggregations(partialBuffer, resultBuffer)
					}

					log.Printf("Reconciler: \n")
					for attrKey, values := range resultBuffer {
						log.Printf("Attribute '%s':", attrKey)
						for value, count := range values {
							log.Printf("'%s': %d occurrences", value, count)
						}
					}
				}
			}
		}
	}()
}

// Stop gracefully shuts down the worker pool
func (s *Service) Stop() {
	close(s.quit)
	close(s.jobsCh)
	s.workersWg.Wait()
	s.flushTicker.Stop()
	// close(s.results)
}

// worker processes jobs from the job queue
func (s *Service) worker(ctx context.Context, workerID int) {
	defer s.workersWg.Done()
	
	// todo: add local map here
	localBuffer := make(domain.AttributeAggregation)

	for {
		select {
		case job, ok := <-s.jobsCh:
			if !ok {
				return // Channel closed, worker should exit
			}
			log.Printf("Worker %d: Processing job", workerID)
			s.mergeAggregations(job, localBuffer)
		case _, ok := <-s.flushCh:
			if !ok {
				return // Channel closed, worker should exit
			}
			log.Printf("Worker %d: Flushing local buffer (%d)", workerID, len(localBuffer))
			select {
			case s.reconcileCh <- localBuffer:
			default:
				// nobody is listening
			}
			localBuffer = make(domain.AttributeAggregation)
		case <-ctx.Done():
			return
		case <-s.quit:
			return
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
