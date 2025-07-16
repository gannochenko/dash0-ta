package attribute_processor

import "time"

type Service struct {
    localBuffers chan map[string]int64
    globalMap    map[string]*int64
    bufferSize   int
    flushTicker  *time.Ticker
}

func New() *Service {
	return &Service{
		localBuffers: make(chan map[string]int64),
		globalMap:    make(map[string]*int64),
		bufferSize:   1000,
		flushTicker:  time.NewTicker(1 * time.Second),
	}
}
