package mediasource

import (
	"github.com/asticode/go-astiav"
	"github.com/harshabose/simple_webrtc_comm/transcode/pkg"
	"github.com/harshabose/tools/buffer/pkg"

	"mediasource/internal"
)

type StreamOption = func(*Stream) error

func WithBufferSize(size int) StreamOption {
	return func(stream *Stream) error {
		stream.buffer = buffer.CreateChannelBuffer(stream.ctx, size, internal.CreateSamplePool())
		return nil
	}
}

func WithDemuxer(containerAddress string, options ...transcode.DemuxerOption) StreamOption {
	return func(stream *Stream) error {
		var err error
		if stream.demuxer, err = transcode.CreateDemuxer(stream.ctx, containerAddress, options...); err != nil {
			return err
		}
		return nil
	}
}

func WithDecoder(options ...transcode.DecoderOption) StreamOption {
	return func(stream *Stream) error {
		var err error
		if stream.decoder, err = transcode.CreateDecoder(stream.ctx, stream.demuxer, options...); err != nil {
			return err
		}
		return nil
	}
}

func WithFilter(filterConfig *transcode.FilterConfig, options ...transcode.FilterOption) StreamOption {
	return func(stream *Stream) error {
		var err error
		if stream.filter, err = transcode.CreateFilter(stream.ctx, stream.decoder, filterConfig, options...); err != nil {
			return err
		}
		return nil
	}
}

func WithEncoder(codec astiav.CodecID, options ...transcode.EncoderOption) StreamOption {
	return func(stream *Stream) error {
		var err error
		if stream.encoder, err = transcode.CreateEncoder(stream.ctx, codec, stream.filter, options...); err != nil {
			return err
		}
		return nil
	}
}
