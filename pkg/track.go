package mediasource

import (
	"context"
	"fmt"

	"github.com/pion/webrtc/v4"
	"github.com/pion/webrtc/v4/pkg/media"
)

// NO BUFFER IMPLEMENTATION

type Track struct {
	track         *webrtc.TrackLocalStaticSample
	rtcCapability *webrtc.RTPCodecCapability
	rtpSender     *webrtc.RTPSender
	stream        *Stream
	priority      Priority
	ctx           context.Context
}

func CreateTrack(ctx context.Context, label string, peerConnection *webrtc.PeerConnection, options ...TrackOption) (*Track, error) {
	var err error
	track := &Track{ctx: ctx, rtcCapability: &webrtc.RTPCodecCapability{}}

	for _, option := range options {
		if err := option(track); err != nil {
			return nil, err
		}
	}

	if track.track, err = webrtc.NewTrackLocalStaticSample(*track.rtcCapability, label, "webrtc"); err != nil {
		return nil, err
	}

	if track.rtpSender, err = peerConnection.AddTrack(track.track); err != nil {
		return nil, err
	}

	fmt.Println("added track to peer connection")

	return track, nil
}

func (track *Track) GetTrack() *webrtc.TrackLocalStaticSample {
	return track.track
}

func (track *Track) GetPriority() Priority {
	return track.priority
}

func (track *Track) Start() {
	if track.track == nil {
		fmt.Println("no remote track set yet. Skipping...")
		return
	}

	track.stream.Start()
	go track.loop()
	fmt.Println("media source track started")
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
			fmt.Println("send samples to remote")
		}
	}
}
