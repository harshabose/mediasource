package mediasource

import (
	"github.com/harshabose/simple_webrtc_comm/transcode/pkg"
	"github.com/harshabose/tools/buffer/pkg"

	"github.com/harshabose/simple_webrtc_comm/mediasource/internal"
)

type StreamOption = func(*Stream) error

func WithBufferSize(size int) StreamOption {
	return func(stream *Stream) error {
		stream.buffer = buffer.CreateChannelBuffer(stream.ctx, size, internal.CreateSamplePool())
		return nil
	}
}

func WithTranscoder(options ...transcode.TranscoderOption) StreamOption {
	return func(stream *Stream) error {
		t, err := transcode.CreateTranscoder(options...)
		if err != nil {
			return err
		}

		stream.transcoder = t
		s, ok := stream.transcoder.(*transcode.Transcoder)
		if !ok {
			return transcode.ErrorInterfaceMismatch
		}
		s.Start()
		return nil
	}
}
