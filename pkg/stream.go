package mediasource

import (
	"context"
	"fmt"
	"time"

	"github.com/asticode/go-astiav"
	"github.com/pion/webrtc/v4/pkg/media"

	"github.com/harshabose/simple_webrtc_comm/transcode/pkg"
	"github.com/harshabose/tools/buffer/pkg"

	"github.com/harshabose/simple_webrtc_comm/mediasource/internal"
)

type Stream struct {
	demuxer *transcode.Demuxer
	decoder *transcode.Decoder
	filter  *transcode.Filter
	encoder *transcode.Encoder
	buffer  buffer.BufferWithGenerator[media.Sample]
	ctx     context.Context
}

func CreateStream(ctx context.Context, options ...StreamOption) (*Stream, error) {
	var (
		err    error
		stream *Stream = &Stream{ctx: ctx}
	)

	for _, option := range options {
		if err = option(stream); err != nil {
			return nil, err
		}
	}

	if stream.buffer == nil {
		stream.buffer = buffer.CreateChannelBuffer(ctx, 256, internal.CreateSamplePool())
	}

	return stream, nil
}

func (stream *Stream) Start() {
	stream.demuxer.Start()
	stream.decoder.Start()
	stream.filter.Start()
	stream.encoder.Start()
	go stream.loop()
	fmt.Println("media source stream started")
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

func (stream *Stream) packetToSample(packet *astiav.Packet) *media.Sample {
	sample := stream.buffer.Generate()

	sample.Data = packet.Data()
	sample.Timestamp = time.Now().UTC()
	sample.Duration = time.Duration(packet.Duration())
	// sample.PacketTimestamp = uint32(float64(packet.Pts()) * (stream.encoder.GetTimeBase().Float64()) * float64(time.Second))
	// sample.PrevDroppedPackets = 0
	// sample.Metadata = nil
	// sample.RTPHeaders = nil

	return sample
}
