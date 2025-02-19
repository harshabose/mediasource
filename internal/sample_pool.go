package internal

import (
	"sync"

	"github.com/harshabose/tools/buffer/pkg"
	"github.com/pion/webrtc/v4/pkg/media"
)

type samplePool struct {
	pool sync.Pool
}

func CreateSamplePool() buffer.Pool[media.Sample] {
	return &samplePool{
		pool: sync.Pool{
			New: func() any {
				return &media.Sample{}
			},
		},
	}
}

func (pool *samplePool) Get() *media.Sample {
	packet, ok := pool.pool.Get().(*media.Sample)

	if packet == nil || !ok {
		return &media.Sample{}
	}
	return packet
}

func (pool *samplePool) Put(sample *media.Sample) {
	if sample == nil {
		return
	}
	pool.pool.Put(sample)
}

func (pool *samplePool) Release() {
	for {
		sample, ok := pool.pool.Get().(*media.Sample)
		if !ok {
			continue
		}
		if sample == nil {
			break
		}

		sample = nil
	}
}
