package mediasource

import "github.com/pion/webrtc/v4"

type Option = func(*Track) error

func WithH264(clockrate uint32) Option {
	return func(track *Track) error {
		var (
			err error = nil
		)
		if track.track, err = webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{
			MimeType:    webrtc.MimeTypeH264,
			ClockRate:   clockrate,
			SDPFmtpLine: "level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=420029",
		}, "video", "webrtc"); err != nil {
			return err
		}
		return nil
	}
}

func WithOpus(samplerate uint32, channelLayout uint16) Option {
	return func(track *Track) error {
		var (
			err error = nil
		)

		if track.track, err = webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{
			MimeType:  webrtc.MimeTypeOpus,
			ClockRate: samplerate,
			Channels:  channelLayout,
		}, "audio", "webrtc"); err != nil {
			return err
		}
		return nil
	}
}
