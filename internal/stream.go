package internal

import (
	"context"
	"fmt"
	"time"

	"github.com/asticode/go-astiav"
	"github.com/harshabose/simple_webrtc_comm/transcode/pkg"
	"github.com/harshabose/tools/buffer/pkg"
	"github.com/pion/webrtc/v4/pkg/media"
)

type Options struct {
	DemuxerOptions []transcode.DemuxerOption
	DecoderOptions []transcode.DecoderOption
	FilterOptions  []transcode.FilterOption
	EncoderOptions []transcode.EncoderOption
}

type Stream struct {
	demuxer *transcode.Demuxer
	decoder *transcode.Decoder
	filter  *transcode.Filter
	encoder *transcode.Encoder
	buffer  buffer.BufferWithGenerator[media.Sample]
	ctx     context.Context
}

func CreateStream(ctx context.Context, containerAddress string, options *Options) (*Stream, error) {
	var (
		demuxer *transcode.Demuxer
		decoder *transcode.Decoder
		filter  *transcode.Filter
		encoder *transcode.Encoder
		err     error
	)

	if demuxer, err = transcode.CreateDemuxer(ctx, containerAddress, options.DemuxerOptions...); err != nil {
		return nil, err
	}
	if decoder, err = transcode.CreateDecoder(ctx, demuxer, append([]transcode.DecoderOption{demuxer.GetDecoderContextOptions()}, options.DecoderOptions...)...); err != nil {
		return nil, err
	}
	if filter, err = transcode.CreateFilter(ctx, decoder, transcode.VideoFilters, decoder.GetSrcFilterContextOptions(), transcode.WithDefaultVideoFilterContentOptions); err != nil {
		return nil, err
	}
	if encoder, err = transcode.CreateEncoder(ctx, filter, transcode.WithLowLatencyVideoEncoderSetting); err != nil {
		return nil, err
	}

	fmt.Println("started encoder with settings:")

	return &Stream{
		demuxer: demuxer,
		decoder: decoder,
		filter:  filter,
		encoder: encoder,
		buffer:  buffer.CreateChannelBuffer(ctx, encoder.GetFPS()*3, CreateSamplePool()),
		ctx:     ctx,
	}, nil
}

func (stream *Stream) Start() {
	stream.demuxer.Start()
	stream.decoder.Start()
	stream.filter.Start()
	stream.encoder.Start()
	go stream.loop()
}

func (stream *Stream) loop() {
	var (
		packet *astiav.Packet
		err    error
	)

	for {
		select {
		case <-stream.ctx.Done():
			return
		case packet = <-stream.encoder.WaitForPacket():
			if err = stream.pushSample(stream.packetToSample(packet)); err != nil {
				stream.encoder.PutBack(packet)
				continue
			}

			stream.encoder.PutBack(packet)
		}
	}
}

func (stream *Stream) pushSample(sample *media.Sample) error {
	ctx, cancel := context.WithTimeout(stream.ctx, time.Second)
	defer cancel()

	return stream.buffer.Push(ctx, sample)
}

func (stream *Stream) GetSample() (*media.Sample, error) {
	ctx, cancel := context.WithTimeout(stream.ctx, time.Second)
	defer cancel()

	return stream.buffer.Pop(ctx)
}

func (stream *Stream) PutBack(sample *media.Sample) {
	stream.buffer.PutBack(sample)
}

func (stream *Stream) WaitForSample() chan *media.Sample {
	return stream.buffer.GetChannel()
}

func (stream *Stream) GetParameterSets() ([]byte, []byte) {
	return stream.encoder.GetSPS(), stream.encoder.GetPPS()
}

func (stream *Stream) packetToSample(packet *astiav.Packet) *media.Sample {
	sample := stream.buffer.Generate()

	sample.Data = packet.Data()
	sample.Timestamp = time.Now().UTC()
	sample.Duration = time.Second / time.Duration(stream.encoder.GetFPS())
	sample.PacketTimestamp = uint32(float64(packet.Pts()) * float64(stream.encoder.GetVideoTimeBase()) * float64(time.Second))
	sample.PrevDroppedPackets = 0
	sample.Metadata = nil
	sample.RTPHeaders = nil

	return sample
}
