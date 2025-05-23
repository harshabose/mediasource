package mediasource

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/asticode/go-astiav"

	"github.com/harshabose/simple_webrtc_comm/transcode/pkg"
)

func TestStream(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s, err := CreateStream(ctx,
		WithDemuxer("0", transcode.WithAvFoundationInputFormatOption),
		WithDecoder(),
		WithFilter(transcode.VideoFilters, transcode.WithVideoScaleFilterContent(1270, 1080), transcode.WithVideoPixelFormatFilterContent(astiav.PixelFormatYuv420P)),
		// WithEncoder(astiav.CodecIDVp8, transcode.WithDefaultVP8Options),
	)
	if err != nil {
		t.FailNow()
	}

	s.Start()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			sample, err := s.GetSample()
			if err != nil {
				t.FailNow()
			}

			fmt.Printf("got a sample:%d @ %s\n", sample.PacketTimestamp, sample.Timestamp.Format(time.RFC822))
		}
	}
}
