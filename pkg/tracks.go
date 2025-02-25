package mediasource

import (
	"context"
	"errors"

	"github.com/pion/interceptor"
	"github.com/pion/webrtc/v4"
)

type Tracks struct {
	bwEstimator         *bandwidthEstimator
	mediaEngine         *webrtc.MediaEngine
	interceptorRegistry *interceptor.Registry
	tracks              map[string]*Track
	ctx                 context.Context
}

func CreateTracks(ctx context.Context, mediaEngine *webrtc.MediaEngine, interceptorRegistry *interceptor.Registry, options ...TracksOption) (*Tracks, error) {
	tracks := &Tracks{
		tracks:              make(map[string]*Track),
		mediaEngine:         mediaEngine,
		interceptorRegistry: interceptorRegistry,
		ctx:                 ctx,
	}

	for _, option := range options {
		if err := option(tracks); err != nil {
			return nil, err
		}
	}

	return tracks, nil
}

func (tracks *Tracks) CreateTrack(peerConnection *webrtc.PeerConnection, options ...TrackOption) error {
	var (
		track *Track
		err   error
	)
	if track, err = CreateTrack(tracks.ctx, peerConnection, append(options, withBandwidthControl(tracks.bwEstimator))...); err != nil {
		return err
	}
	if _, exists := tracks.tracks[track.track.ID()]; exists {
		return errors.New("track already exists")
	}
	tracks.tracks[track.track.ID()] = track
	return nil
}

func (tracks *Tracks) StartTrack(id string) {
	if track, ok := tracks.tracks[id]; ok {
		track.Start()
	}
	if tracks.bwEstimator != nil {
		tracks.bwEstimator.Start()
	}
}

func (tracks *Tracks) StartAll() {
	for _, track := range tracks.tracks {
		tracks.StartTrack(track.track.ID())
	}
}
