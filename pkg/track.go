package mediasource

import (
	"context"
	"fmt"
	"time"

	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v4"
	"github.com/pion/webrtc/v4/pkg/media"
)

// NO BUFFER IMPLEMENTATION

type Track struct {
	track         *webrtc.TrackLocalStaticSample
	rtcCapability *webrtc.RTPCodecCapability
	rtpSender     *webrtc.RTPSender
	stream        *Stream
	priority      Priority
	ctx           context.Context
}

func CreateTrack(ctx context.Context, label string, peerConnection *webrtc.PeerConnection, options ...TrackOption) (*Track, error) {
	var err error
	track := &Track{ctx: ctx, rtcCapability: &webrtc.RTPCodecCapability{}}

	for _, option := range options {
		if err := option(track); err != nil {
			return nil, err
		}
	}

	if track.track, err = webrtc.NewTrackLocalStaticSample(*track.rtcCapability, label, "webrtc"); err != nil {
		return nil, err
	}

	fmt.Println("rtc capability:", track.track.Codec().MimeType)

	if track.rtpSender, err = peerConnection.AddTrack(track.track); err != nil {
		return nil, err
	}

	fmt.Println("added track to peer connection")

	return track, nil
}

func (track *Track) GetTrack() *webrtc.TrackLocalStaticSample {
	return track.track
}

func (track *Track) GetPriority() Priority {
	return track.priority
}

func (track *Track) Start() {
	if track.track == nil {
		fmt.Println("no remote track set yet. Skipping...")
		return
	}

	track.stream.Start()
	go track.loop()
	go track.rtpSenderLoop()
	fmt.Println("media source track started")
}

func (track *Track) loop() {
	var (
		sample *media.Sample = nil
		err    error         = nil
	)

loop:
	for {
		select {
		case <-track.ctx.Done():
			return
		case sample = <-track.stream.WaitForSample():
			if err = track.track.WriteSample(*sample); err != nil {
				fmt.Printf("Error pushing packet: %v\n", err)
				track.stream.PutBack(sample)
				continue loop
			}
			track.stream.PutBack(sample)
		}
	}
}

func (track *Track) rtpSenderLoop() {
	readRTCPWithAnalysis(track.rtpSender)
}

// Enhanced RTCP reader that parses and analyzes TWCC feedback
func readRTCPWithAnalysis(rtpSender *webrtc.RTPSender) {
	rtcpBuf := make([]byte, 1500)
	lastTWCCReport := time.Now()

	for {
		n, _, rtcpErr := rtpSender.Read(rtcpBuf)
		if rtcpErr != nil {
			fmt.Printf("RTCP read error: %v\n", rtcpErr)
			return
		}

		// Parse the RTCP packet
		packets, err := rtcp.Unmarshal(rtcpBuf[:n])
		if err != nil {
			fmt.Printf("Failed to unmarshal RTCP: %v\n", err)
			continue
		}

		// Analyze each RTCP packet
		for _, packet := range packets {
			switch p := packet.(type) {
			case *rtcp.TransportLayerCC:
				// TWCC feedback packet - this is what we're interested in!
				analyzeTWCCFeedback(p)
				lastTWCCReport = time.Now()

			case *rtcp.ReceiverReport:
				analyzeReceiverReport(p)

			case *rtcp.SenderReport:
				analyzeSenderReport(p)

			case *rtcp.ReceiverEstimatedMaximumBitrate:
				// REMB packet
				analyzeREMBFeedback(p)

			default:
				fmt.Printf("📄 Other RTCP packet: %T\n", packet)
			}
		}

		// Check if we're missing TWCC feedback
		if time.Since(lastTWCCReport) > 5*time.Second {
			fmt.Printf("⚠️  No TWCC feedback received for %v - BWE may be stuck!\n",
				time.Since(lastTWCCReport))
		}
	}
}

func analyzeTWCCFeedback(twcc *rtcp.TransportLayerCC) {
	fmt.Printf("📈 TWCC Feedback Received:\n")
	fmt.Printf("  Media SSRC: %d\n", twcc.MediaSSRC)
	fmt.Printf("  Feedback Packet Count: %d\n", twcc.FbPktCount)
	fmt.Printf("  Packet Status Count: %d\n", twcc.PacketStatusCount)
	fmt.Printf("  Reference Time: %d\n", twcc.ReferenceTime)

	// Count received packets
	receivedCount := 0
	lostCount := 0

	for i, _ := range twcc.PacketChunks {
		if i >= int(twcc.PacketStatusCount) {
			break
		}
	}

	totalPackets := receivedCount + lostCount
	packetLossRate := float64(lostCount) / float64(totalPackets) * 100

	fmt.Printf("  Packets: %d received, %d lost (%.1f%% loss)\n",
		receivedCount, lostCount, packetLossRate)

	// Analyze receive deltas for jitter
	if len(twcc.RecvDeltas) > 0 {
		analyzeReceiveDeltas(twcc.RecvDeltas)
	}

	fmt.Println("  ----------------------------------------")
}

func analyzeReceiveDeltas(deltas []*rtcp.RecvDelta) {
	if len(deltas) < 2 {
		return
	}

	// Calculate jitter and timing information
	var totalDelta time.Duration
	var maxDelta, minDelta time.Duration = 0, time.Hour

	for _, delta := range deltas {
		deltaDuration := delta.Delta
		totalDelta += time.Duration(deltaDuration)

		if time.Duration(deltaDuration) > (maxDelta) {
			maxDelta = time.Duration(deltaDuration)
		}
		if time.Duration(deltaDuration) < minDelta {
			minDelta = time.Duration(deltaDuration)
		}
	}

	avgDelta := totalDelta / time.Duration(len(deltas))
	jitter := maxDelta - minDelta

	fmt.Printf("  Timing Analysis:\n")
	fmt.Printf("    Avg Delta: %v\n", avgDelta)
	fmt.Printf("    Jitter: %v (max: %v, min: %v)\n", jitter, maxDelta, minDelta)

	// Warning for high jitter
	if jitter > 50*time.Millisecond {
		fmt.Printf("    ⚠️  High jitter detected: %v\n", jitter)
	}
}

func analyzeReceiverReport(rr *rtcp.ReceiverReport) {
	fmt.Printf("📊 Receiver Report:\n")
	fmt.Printf("  SSRC: %d\n", rr.SSRC)

	for _, report := range rr.Reports {
		fmt.Printf("  Report for SSRC %d:\n", report.SSRC)
		fmt.Printf("    Fraction Lost: %d/256 (%.1f%%)\n",
			report.FractionLost, float64(report.FractionLost)/256*100)
		fmt.Printf("    Total Lost: %d packets\n", report.TotalLost)
		fmt.Printf("    Highest Seq: %d\n", report.LastSequenceNumber)
		fmt.Printf("    Jitter: %d timestamp units\n", report.Jitter)
	}
	fmt.Println("  ----------------------------------------")
}

func analyzeSenderReport(sr *rtcp.SenderReport) {
	fmt.Printf("📤 Sender Report:\n")
	fmt.Printf("  SSRC: %d\n", sr.SSRC)
	fmt.Printf("  NTP Time: %d\n", sr.NTPTime)
	fmt.Printf("  RTP Time: %d\n", sr.RTPTime)
	fmt.Printf("  Packet Count: %d\n", sr.PacketCount)
	fmt.Printf("  Octet Count: %d\n", sr.OctetCount)
	fmt.Println("  ----------------------------------------")
}

func analyzeREMBFeedback(remb *rtcp.ReceiverEstimatedMaximumBitrate) {
	fmt.Printf("📶 REMB Feedback:\n")
	fmt.Printf("  Sender SSRC: %d\n", remb.SenderSSRC)
	fmt.Printf("  Bitrate: %.2f Mbps\n", float64(remb.Bitrate)/1_000_000)
	fmt.Printf("  Media SSRCs: %v\n", remb.SSRCs)
	fmt.Println("  ----------------------------------------")
}
