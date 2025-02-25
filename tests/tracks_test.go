package tests

import (
	"context"
	"testing"
	"time"

	"github.com/asticode/go-astiav"
	"github.com/harshabose/simple_webrtc_comm/transcode/pkg"
	"github.com/pion/interceptor"
	"github.com/pion/webrtc/v4"

	"mediasource/pkg"
)

func TestTracks(t *testing.T) {
	ctx := context.Background()

	mediaEngine := &webrtc.MediaEngine{}
	registry := &interceptor.Registry{}

	tracks, err := mediasource.CreateTracks(ctx, mediaEngine, registry,
		mediasource.WithH264MediaEngine(90000, mediasource.PacketisationMode1, mediasource.ProfileLevelBaseline42),
		mediasource.WithOpusMediaEngine(48000, 2, mediasource.Stereo),
		mediasource.WithNACKInterceptor(mediasource.NACKGeneratorDefault, mediasource.NACKResponderDefault),
		mediasource.WithRTCPReportsInterceptor(mediasource.RTCPReportIntervalDefault),
		mediasource.WithJitterBufferInterceptor(),
		mediasource.WithFLEXFECInterceptor(),
		mediasource.WithTWCCSenderInterceptor(mediasource.TWCCIntervalDefault),
		mediasource.WithBandwidthEstimatorInterceptor(8000, 100*time.Millisecond),
	)
	if err != nil {
		t.Error(err)
	}

	api := webrtc.NewAPI(webrtc.WithMediaEngine(mediaEngine), webrtc.WithInterceptorRegistry(registry))
	peerConnection, err := api.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		t.Error(err)
	}

	if err := tracks.CreateTrack(peerConnection,
		mediasource.WithH264Track(90000, "test-track"),
		mediasource.WithPriority(mediasource.Level3),
		mediasource.WithStream(
			mediasource.WithDemuxer("/dev/video0"),
			mediasource.WithDecoder(),
			mediasource.WithFilter(transcode.VideoFilters),
			mediasource.WithEncoder(astiav.CodecIDH264),
		),
	); err != nil {
		t.Error(err)
	}

	tracks.StartTrack("test-track")
}
