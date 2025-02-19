package mediasource

import (
	"context"
	"fmt"
	"time"

	"github.com/pion/webrtc/v4"
	"github.com/pion/webrtc/v4/pkg/media"

	"mediasource/internal"
)

// NO BUFFER IMPLEMENTATION

type Track struct {
	track  *webrtc.TrackLocalStaticSample
	stream *internal.Stream
	ctx    context.Context
}

func CreateLocalTrack(ctx context.Context, stream *internal.Stream, options ...Option) (*Track, error) {
	track := &Track{stream: stream, ctx: ctx}

	for _, option := range options {
		if err := option(track); err != nil {
			return nil, err
		}
	}

	return track, nil
}

func (track *Track) GetTrack() *webrtc.TrackLocalStaticSample {
	return track.track
}

func (track *Track) Start() {
	if track.track == nil {
		fmt.Printf("no remote track set yet. Skipping...")
		return
	}

	track.stream.Start()
	go track.loop()
}

func (track *Track) loop() {
	var (
		sample *media.Sample = nil
		err    error         = nil
		ticker *time.Ticker  = nil
	)

	ticker = time.NewTicker(time.Second)
	defer ticker.Stop()

loop:
	for {
		select {
		case <-track.ctx.Done():
			return
		case sample = <-track.stream.WaitForSample():
			if err = track.track.WriteSample(*sample); err != nil {
				fmt.Printf("Error pushing packet: %v\n", err)
				track.stream.PutBack(sample)
				continue loop
			}
			track.stream.PutBack(sample)
		}
	}
}
