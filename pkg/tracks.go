package mediasource

import (
	"context"
	"errors"

	"github.com/pion/webrtc/v4"
)

type Tracks struct {
	tracks map[string]*Track
	ctx    context.Context
}

func CreateTracks(ctx context.Context, options ...TracksOption) (*Tracks, error) {
	tracks := &Tracks{
		tracks: make(map[string]*Track),
		ctx:    ctx,
	}

	for _, option := range options {
		if err := option(tracks); err != nil {
			return nil, err
		}
	}

	return tracks, nil
}

func (tracks *Tracks) CreateTrack(label string, peerConnection *webrtc.PeerConnection, options ...TrackOption) (*Track, error) {
	var (
		track *Track
		err   error
	)

	if track, err = CreateTrack(tracks.ctx, label, peerConnection, options...); err != nil {
		return nil, err
	}
	if _, exists := tracks.tracks[track.track.ID()]; exists {
		return nil, errors.New("track already exists")
	}
	tracks.tracks[track.track.ID()] = track
	return track, nil
}

func (tracks *Tracks) GetTrack(id string) (*Track, error) {
	track, exists := tracks.tracks[id]
	if !exists {
		return nil, errors.New("track does not exits")
	}

	return track, nil
}

func (tracks *Tracks) StartTrack(id string) {
	if track, ok := tracks.tracks[id]; ok {
		track.Start()
	}
}

func (tracks *Tracks) StartAll() {
	for _, track := range tracks.tracks {
		tracks.StartTrack(track.track.ID())
	}
}
