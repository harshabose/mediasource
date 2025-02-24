package mediasource

import (
	"context"
	"errors"
	"github.com/pion/interceptor/pkg/cc"
	"sync"
	"time"
)

type consumer struct {
	channel chan int64
	track   *Track
}

type bandwidthEstimator struct {
	estimator cc.BandwidthEstimator
	consumers map[string]*consumer
	interval  time.Duration
	mutex     sync.RWMutex
	ctx       context.Context
}

func (be *bandwidthEstimator) Start() {
	go be.loop()
}

func (be *bandwidthEstimator) SetConsumer(id string, setChannel func(chan int64), track *Track) error {
	be.mutex.Lock()
	defer be.mutex.Unlock()

	if _, exits := be.consumers[id]; exits {
		return errors.New("consumer already exists")
	}

	be.consumers[id] = &consumer{channel: make(chan int64), track: track}
	setChannel(be.consumers[id].channel)

	return nil
}

func (be *bandwidthEstimator) loop() {
	// wait here
	for {
		be.mutex.RLock()

		select {
		case <-be.ctx.Done():
			return
		default:
			be.estimate()
		}

		be.mutex.RUnlock()
	}
}

func (be *bandwidthEstimator) estimate() {
	var totalPriority Priority

	if len(be.consumers) == 0 {
		return
	}

	for _, consumer := range be.consumers {
		totalPriority += consumer.track.priority
	}

	totalBitrate := be.estimator.GetTargetBitrate()

	for _, consumer := range be.consumers {
		if consumer.track.priority == Level0 {
			continue
		}
		select {
		case consumer.channel <- int64(float64(totalBitrate) * float64(consumer.track.priority) / float64(totalPriority)):
		}
	}
}
