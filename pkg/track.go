package mediasource

import (
	"context"
	"fmt"

	"github.com/pion/webrtc/v4"
	"github.com/pion/webrtc/v4/pkg/media"
)

// NO BUFFER IMPLEMENTATION

type Track struct {
	track     *webrtc.TrackLocalStaticSample
	rtpSender *webrtc.RTPSender
	stream    *Stream
	priority  Priority
	ctx       context.Context
}

func CreateTrack(ctx context.Context, peerConnection *webrtc.PeerConnection, options ...TrackOption) (*Track, error) {
	var err error
	track := &Track{ctx: ctx}

	for _, option := range options {
		if err := option(track); err != nil {
			return nil, err
		}
	}

	if track.rtpSender, err = peerConnection.AddTrack(track.track); err != nil {
		return nil, err
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
	fmt.Printf("media source track started")
}

func (track *Track) loop() {
	var (
		sample *media.Sample = nil
		err    error         = nil
	)

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
