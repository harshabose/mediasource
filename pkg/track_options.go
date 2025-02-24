package mediasource

import (
	"github.com/pion/webrtc/v4"
)

type TrackOption = func(*Track) error

func WithH264Track(clockrate uint32, id string) TrackOption {
	return func(track *Track) error {
		var (
			err error = nil
		)

		if track.track, err = webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{
			MimeType:    webrtc.MimeTypeH264,
			ClockRate:   clockrate,
			SDPFmtpLine: "level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=420029",
		}, id, "webrtc"); err != nil {
			return err
		}
		return nil
	}
}

func WithOpusTrack(samplerate uint32, channelLayout uint16, id string) TrackOption {
	return func(track *Track) error {
		var (
			err error = nil
		)

		if track.track, err = webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{
			MimeType:  webrtc.MimeTypeOpus,
			ClockRate: samplerate,
			Channels:  channelLayout,
		}, id, "webrtc"); err != nil {
			return err
		}
		return nil
	}
}

func WithStreamOptions(options ...StreamOption) TrackOption {
	return func(track *Track) error {
		for _, option := range options {
			if err := option(track.stream); err != nil {
				return err
			}
		}
		return nil
	}
}

func WithPriority(level Priority) TrackOption {
	return func(track *Track) error {
		track.priority = level
		return nil
	}
}

type Priority uint8

const (
	Level0 Priority = 0
	Level1 Priority = 1
	Level2 Priority = 2
	Level3 Priority = 3
	Level4 Priority = 4
	Level5 Priority = 5
)

func withBandwidthControl(estimator *bandwidthEstimator) TrackOption {
	return func(track *Track) error {
		return estimator.SetConsumer(track.track.ID(), track.stream.encoder.SetBitrateChannel, track)
	}
}
