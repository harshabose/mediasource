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
	transcoder transcode.CanProduceMediaPacket
	buffer     buffer.BufferWithGenerator[media.Sample]

	frameCount int
	startTime  time.Time

	ctx context.Context
}

func CreateStream(ctx context.Context, options ...StreamOption) (*Stream, error) {
	var (
		err    error
		stream = &Stream{ctx: ctx}
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
	go stream.loop()
	fmt.Println("media source stream started")
}

func (stream *Stream) loop() {
	var (
		packet *astiav.Packet
		err    error
	)
	ticker := time.NewTicker(40 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-stream.ctx.Done():
			return
		case <-ticker.C:
			packet, err = stream.getPacket()
			if err != nil {
				// fmt.Println("unable to get packet from transcoder; err:", err.Error())
				continue
			}
			if err = stream.pushSample(stream.packetToSample(packet)); err != nil {
				stream.transcoder.PutBack(packet)
				continue
			}

			stream.transcoder.PutBack(packet)
		}
	}
}

func (stream *Stream) pushSample(sample *media.Sample) error {
	if sample == nil {
		fmt.Println("got nil sample skipping")
		return nil
	}
	ctx, cancel := context.WithTimeout(stream.ctx, time.Second)
	defer cancel()

	return stream.buffer.Push(ctx, sample)
}

func (stream *Stream) getPacket() (*astiav.Packet, error) {
	ctx, cancel := context.WithTimeout(stream.ctx, 50*time.Millisecond)
	defer cancel()

	return stream.transcoder.GetPacket(ctx)
}

func (stream *Stream) GetSample() (*media.Sample, error) {
	ctx, cancel := context.WithTimeout(stream.ctx, 50*time.Millisecond)
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
	if packet == nil {
		fmt.Println("🚨 ERROR: Received nil packet")
		return nil
	}

	sample := stream.buffer.Generate()
	sample.Data = packet.Data()

	// Initialize on first packet
	if stream.frameCount == 0 {
		stream.startTime = time.Now().UTC()
	}

	stream.frameCount++

	// Calculate expected timestamp for this frame number
	frameDuration := time.Second / time.Duration(25)
	expectedTimestamp := stream.startTime.Add(time.Duration(stream.frameCount-1) * frameDuration)

	// Use expected timestamp (not real-time) for consistent spacing
	sample.Timestamp = expectedTimestamp
	sample.Duration = frameDuration

	return sample
}
