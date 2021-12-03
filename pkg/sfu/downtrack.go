package sfu

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/pion/rtcp"
	"github.com/pion/rtp"
	"github.com/pion/sdp/v3"
	"github.com/pion/transport/packetio"
	"github.com/pion/webrtc/v3"

	"github.com/livekit/livekit-server/pkg/sfu/buffer"
)

// TrackSender defines a  interface send media to remote peer
type TrackSender interface {
	UptrackLayersChange(availableLayers []uint16)
	WriteRTP(p *buffer.ExtPacket, layer int32) error
	Close()
	// ID is the globally unique identifier for this Track.
	ID() string
	SetTrackType(isSimulcast bool)
	Codec() webrtc.RTPCodecCapability
	PeerID() string
}

// DownTrackType determines the type of track
type DownTrackType int

const (
	SimpleDownTrack DownTrackType = iota + 1
	SimulcastDownTrack
)

const (
	RTPPaddingMaxPayloadSize      = 255
	RTPPaddingEstimatedHeaderSize = 20
	RTPBlankFramesMax             = 6

	InvalidSpatialLayer  = -1
	InvalidTemporalLayer = -1
)

type SequenceNumberOrdering int

const (
	SequenceNumberOrderingContiguous SequenceNumberOrdering = iota
	SequenceNumberOrderingOutOfOrder
	SequenceNumberOrderingGap
	SequenceNumberOrderingDuplicate
)

type ForwardingStatus int

const (
	ForwardingStatusOff ForwardingStatus = iota
	ForwardingStatusPartial
	ForwardingStatusOptimal
)

var (
	ErrUnknownKind                       = errors.New("unknown kind of codec")
	ErrOutOfOrderSequenceNumberCacheMiss = errors.New("out-of-order sequence number not found in cache")
	ErrPaddingOnlyPacket                 = errors.New("padding only packet that need not be forwarded")
	ErrDuplicatePacket                   = errors.New("duplicate packet")
	ErrPaddingNotOnFrameBoundary         = errors.New("padding cannot send on non-frame boundary")
	ErrNotVP8                            = errors.New("not VP8")
	ErrOutOfOrderVP8PictureIdCacheMiss   = errors.New("out-of-order VP8 picture id not found in cache")
	ErrFilteredVP8TemporalLayer          = errors.New("filtered VP8 temporal layer")
	ErrNoRequiredBuff                    = errors.New("buff size if less than required")
)

var (
	VP8KeyFrame1x1 = []byte{0x10, 0x02, 0x00, 0x9d, 0x01, 0x2a, 0x01, 0x00, 0x01, 0x00, 0x0b, 0xc7, 0x08, 0x85, 0x85, 0x88, 0x85, 0x84, 0x88, 0x3f, 0x82, 0x00, 0x0c, 0x0d, 0x60, 0x00, 0xfe, 0xe6, 0xb5, 0x00}

	H264KeyFrame2x2SPS = []byte{0x67, 0x42, 0xc0, 0x1f, 0x0f, 0xd9, 0x1f, 0x88, 0x88, 0x84, 0x00, 0x00, 0x03, 0x00, 0x04, 0x00, 0x00, 0x03, 0x00, 0xc8, 0x3c, 0x60, 0xc9, 0x20}
	H264KeyFrame2x2PPS = []byte{0x68, 0x87, 0xcb, 0x83, 0xcb, 0x20}
	H264KeyFrame2x2IDR = []byte{0x65, 0x88, 0x84, 0x0a, 0xf2, 0x62, 0x80, 0x00, 0xa7, 0xbe}

	H264KeyFrame2x2 = [][]byte{H264KeyFrame2x2SPS, H264KeyFrame2x2PPS, H264KeyFrame2x2IDR}
)

type TranslationParamsRTP struct {
	snOrdering     SequenceNumberOrdering
	sequenceNumber uint16
	timestamp      uint32
}

type TranslationParamsVP8 struct {
	header *buffer.VP8
}

type TranslationParams struct {
	shouldDrop    bool
	shouldSendPLI bool
	rtp           *TranslationParamsRTP
	vp8           *TranslationParamsVP8
}

type SnTs struct {
	sequenceNumber uint16
	timestamp      uint32
}

type VideoLayers struct {
	spatial  int32
	temporal int32
}

type ReceiverReportListener func(dt *DownTrack, report *rtcp.ReceiverReport)

// DownTrack  implements TrackLocal, is the track used to write packets
// to SFU Subscriber, the track handle the packets for simple, simulcast
// and SVC Publisher.
type DownTrack struct {
	id            string
	peerID        string
	bound         atomicBool
	kind          webrtc.RTPCodecType
	mime          string
	ssrc          uint32
	streamID      string
	maxTrack      int
	payloadType   uint8
	sequencer     *sequencer
	trackType     DownTrackType
	bufferFactory *buffer.Factory
	payload       *[]byte

	forwarder *Forwarder

	codec                   webrtc.RTPCodecCapability
	rtpHeaderExtensions     []webrtc.RTPHeaderExtensionParameter
	receiver                TrackReceiver
	transceiver             *webrtc.RTPTransceiver
	writeStream             webrtc.TrackLocalWriter
	onCloseHandler          func()
	onBind                  func()
	receiverReportListeners []ReceiverReportListener
	listenerLock            sync.RWMutex
	closeOnce               sync.Once

	// Report helpers
	octetCount   atomicUint32
	packetCount  atomicUint32
	lossFraction atomicUint8

	// Debug info
	lastPli     atomicInt64
	lastRTP     atomicInt64
	pktsDropped atomicUint32

	// RTCP callbacks
	onRTCP func([]rtcp.Packet)
	onREMB func(dt *DownTrack, remb *rtcp.ReceiverEstimatedMaximumBitrate)

	// simulcast layer availability change callback
	onAvailableLayersChanged func(dt *DownTrack)

	// subscription change callback
	onSubscriptionChanged func(dt *DownTrack)

	// max layer change callback
	onSubscribedLayersChanged func(dt *DownTrack, layers VideoLayers)

	// packet sent callback
	onPacketSent []func(dt *DownTrack, size int)
}

// NewDownTrack returns a DownTrack.
func NewDownTrack(c webrtc.RTPCodecCapability, r TrackReceiver, bf *buffer.Factory, peerID string, mt int) (*DownTrack, error) {
	var kind webrtc.RTPCodecType
	switch {
	case strings.HasPrefix(c.MimeType, "audio/"):
		kind = webrtc.RTPCodecTypeAudio
	case strings.HasPrefix(c.MimeType, "video/"):
		kind = webrtc.RTPCodecTypeVideo
	default:
		kind = webrtc.RTPCodecType(0)
	}

	d := &DownTrack{
		id:            r.TrackID(),
		peerID:        peerID,
		maxTrack:      mt,
		streamID:      r.StreamID(),
		bufferFactory: bf,
		receiver:      r,
		codec:         c,
		kind:          kind,
		forwarder:     NewForwarder(c, kind),
	}

	if strings.ToLower(c.MimeType) == "video/vp8" {
		d.payload = PacketFactory.Get().(*[]byte)
	}

	return d, nil
}

func (d *DownTrack) SetTrackType(isSimulcast bool) {
	if isSimulcast {
		d.trackType = SimulcastDownTrack
	} else {
		d.trackType = SimpleDownTrack
	}
}

// Bind is called by the PeerConnection after negotiation is complete
// This asserts that the code requested is supported by the remote peer.
// If so it setups all the state (SSRC and PayloadType) to have a call
func (d *DownTrack) Bind(t webrtc.TrackLocalContext) (webrtc.RTPCodecParameters, error) {
	parameters := webrtc.RTPCodecParameters{RTPCodecCapability: d.codec}
	if codec, err := codecParametersFuzzySearch(parameters, t.CodecParameters()); err == nil {
		d.ssrc = uint32(t.SSRC())
		d.payloadType = uint8(codec.PayloadType)
		d.writeStream = t.WriteStream()
		d.mime = strings.ToLower(codec.MimeType)
		if rr := d.bufferFactory.GetOrNew(packetio.RTCPBufferPacket, uint32(t.SSRC())).(*buffer.RTCPReader); rr != nil {
			rr.OnPacket(func(pkt []byte) {
				d.handleRTCP(pkt)
			})
		}
		if strings.HasPrefix(d.codec.MimeType, "video/") {
			d.sequencer = newSequencer(d.maxTrack)
		}
		if d.onBind != nil {
			d.onBind()
		}
		d.bound.set(true)
		return codec, nil
	}
	return webrtc.RTPCodecParameters{}, webrtc.ErrUnsupportedCodec
}

// Unbind implements the teardown logic when the track is no longer needed. This happens
// because a track has been stopped.
func (d *DownTrack) Unbind(_ webrtc.TrackLocalContext) error {
	d.bound.set(false)
	d.receiver.DeleteDownTrack(d.peerID)
	return nil
}

// ID is the unique identifier for this Track. This should be unique for the
// stream, but doesn't have to globally unique. A common example would be 'audio' or 'video'
// and StreamID would be 'desktop' or 'webcam'
func (d *DownTrack) ID() string { return d.id }

// Codec returns current track codec capability
func (d *DownTrack) Codec() webrtc.RTPCodecCapability { return d.codec }

// StreamID is the group this track belongs too. This must be unique
func (d *DownTrack) StreamID() string { return d.streamID }

func (d *DownTrack) PeerID() string { return d.peerID }

// Sets RTP header extensions for this track
func (d *DownTrack) SetRTPHeaderExtensions(rtpHeaderExtensions []webrtc.RTPHeaderExtensionParameter) {
	d.rtpHeaderExtensions = rtpHeaderExtensions
}

// Kind controls if this TrackLocal is audio or video
func (d *DownTrack) Kind() webrtc.RTPCodecType {
	return d.kind
}

func (d *DownTrack) SSRC() uint32 {
	return d.ssrc
}

func (d *DownTrack) Stop() error {
	if d.transceiver != nil {
		return d.transceiver.Stop()
	}
	return fmt.Errorf("d.transceiver not exists")
}

func (d *DownTrack) SetTransceiver(transceiver *webrtc.RTPTransceiver) {
	d.transceiver = transceiver
}

// WriteRTP writes a RTP Packet to the DownTrack
func (d *DownTrack) WriteRTP(extPkt *buffer.ExtPacket, layer int32) error {
	d.lastRTP.set(time.Now().UnixNano())

	if !d.bound.get() {
		return nil
	}

	tp, err := d.forwarder.GetTranslationParams(extPkt, layer)
	if tp.shouldSendPLI {
		d.lastPli.set(time.Now().UnixNano())
		d.receiver.SendPLI(layer)
	}
	if tp.shouldDrop {
		d.pktsDropped.add(1)
		return err
	}

	payload := extPkt.Packet.Payload
	if tp.vp8 != nil {
		incomingVP8, _ := extPkt.Payload.(buffer.VP8)
		payload, err = d.translateVP8Packet(&extPkt.Packet, &incomingVP8, tp.vp8.header)
		if err != nil {
			d.pktsDropped.add(1)
			return err
		}
	}

	if d.sequencer != nil {
		meta := d.sequencer.push(extPkt.Packet.SequenceNumber, tp.rtp.sequenceNumber, tp.rtp.timestamp, 0, extPkt.Head)
		if meta != nil && tp.vp8 != nil {
			meta.packVP8(tp.vp8.header)
		}
	}

	hdr, err := d.getTranslatedRTPHeader(extPkt, tp.rtp)
	if err != nil {
		d.pktsDropped.add(1)
		return err
	}

	_, err = d.writeStream.WriteRTP(hdr, payload)
	if err == nil {
		for _, f := range d.onPacketSent {
			f(d, hdr.MarshalSize()+len(payload))
		}
	} else {
		d.pktsDropped.add(1)
	}

	// LK-TODO maybe include RTP header size also
	d.UpdateStats(uint32(len(payload)))

	return err
}

// WritePaddingRTP tries to write as many padding only RTP packets as necessary
// to satisfy given size to the DownTrack
func (d *DownTrack) WritePaddingRTP(bytesToSend int) int {
	if d.packetCount.get() == 0 {
		return 0
	}

	// LK-TODO-START
	// Ideally should look at header extensions negotiated for
	// track and decide if padding can be sent. But, browsers behave
	// in unexpected ways when using audio for bandwidth estimation and
	// padding is mainly used to probe for excess available bandwidth.
	// So, to be safe, limit to video tracks
	// LK-TODO-END
	if d.kind == webrtc.RTPCodecTypeAudio {
		return 0
	}

	// LK-TODO-START
	// Potentially write padding even if muted. Given that padding
	// can be sent only on frame boudaries, writing on disabled tracks
	// will give more options.
	// LK-TODO-END
	if d.forwarder.Muted() {
		return 0
	}

	// RTP padding maximum is 255 bytes. Break it up.
	// Use 20 byte as estimate of RTP header size (12 byte header + 8 byte extension)
	num := (bytesToSend + RTPPaddingMaxPayloadSize + RTPPaddingEstimatedHeaderSize - 1) / (RTPPaddingMaxPayloadSize + RTPPaddingEstimatedHeaderSize)
	if num == 0 {
		return 0
	}

	snts, err := d.forwarder.GetSnTsForPadding(num)
	if err != nil {
		return 0
	}

	// LK-TODO Look at load balancing a la sfu.Receiver to spread across available CPUs
	bytesSent := 0
	for i := 0; i < len(snts); i++ {
		// LK-TODO-START
		// Hold sending padding packets till first RTCP-RR is received for this RTP stream.
		// That is definitive proof that the remote side knows about this RTP stream.
		// The packet count check at the beginning of this function gates sending padding
		// on as yet unstarted streams which is a reasonble check.
		// LK-TODO-END

		hdr := rtp.Header{
			Version:        2,
			Padding:        true,
			Marker:         false,
			PayloadType:    d.payloadType,
			SequenceNumber: snts[i].sequenceNumber,
			Timestamp:      snts[i].timestamp,
			SSRC:           d.ssrc,
			CSRC:           []uint32{},
		}

		err = d.writeRTPHeaderExtensions(&hdr)
		if err != nil {
			return bytesSent
		}

		payload := make([]byte, RTPPaddingMaxPayloadSize)
		// last byte of padding has padding size including that byte
		payload[RTPPaddingMaxPayloadSize-1] = byte(RTPPaddingMaxPayloadSize)

		_, err = d.writeStream.WriteRTP(&hdr, payload)
		if err != nil {
			return bytesSent
		}

		// LK-TODO - check if we should keep separate padding stats
		size := hdr.MarshalSize() + len(payload)
		d.UpdateStats(uint32(size))

		// LK-TODO-START
		// NACK buffer for these probe packets.
		// Probably okay to absorb the NACKs for these and ignore them.
		// Retransmssion is probably a sign of network congestion/badness.
		// So, retransmitting padding packets is only going to make matters worse.
		// LK-TODO-END

		bytesSent += size
	}

	return bytesSent
}

// Mute enables or disables media forwarding
func (d *DownTrack) Mute(val bool) {
	changed := d.forwarder.Mute(val)
	if !changed {
		return
	}

	if val {
		d.lossFraction.set(0)
	}

	if d.onSubscriptionChanged != nil {
		d.onSubscriptionChanged(d)
	}
}

// Close track
func (d *DownTrack) Close() {
	d.forwarder.Mute(true)

	// write blank frames after disabling so that other frames do not interfere.
	// Idea here is to send blank 1x1 key frames to flush the decoder buffer at the remote end.
	// Otherwise, with transceiver re-use last frame from previous stream is held in the
	// display buffer and there could be a brief moment where the previous stream is displayed.
	d.writeBlankFrameRTP()

	d.closeOnce.Do(func() {
		Logger.V(1).Info("Closing sender", "peer_id", d.peerID, "kind", d.kind)
		if d.payload != nil {
			PacketFactory.Put(d.payload)
		}
		if d.onCloseHandler != nil {
			d.onCloseHandler()
		}
	})
}

func (d *DownTrack) SetMaxSpatialLayer(spatialLayer int32) {
	changed, maxLayers := d.forwarder.SetMaxSpatialLayer(spatialLayer)
	if !changed {
		return
	}

	if d.onSubscribedLayersChanged != nil {
		d.onSubscribedLayersChanged(d, maxLayers)
	}
}

func (d *DownTrack) SetMaxTemporalLayer(temporalLayer int32) {
	changed, maxLayers := d.forwarder.SetMaxTemporalLayer(temporalLayer)
	if !changed {
		return
	}

	if d.onSubscribedLayersChanged != nil {
		d.onSubscribedLayersChanged(d, maxLayers)
	}
}

func (d *DownTrack) MaxLayers() VideoLayers {
	return d.forwarder.MaxLayers()
}

func (d *DownTrack) GetForwardingStatus() ForwardingStatus {
	return d.forwarder.GetForwardingStatus()
}

func (d *DownTrack) UptrackLayersChange(availableLayers []uint16) {
	d.forwarder.UptrackLayersChange(availableLayers)

	if d.onAvailableLayersChanged != nil {
		d.onAvailableLayersChanged(d)
	}
}

// OnCloseHandler method to be called on remote tracked removed
func (d *DownTrack) OnCloseHandler(fn func()) {
	d.onCloseHandler = fn
}

func (d *DownTrack) OnBind(fn func()) {
	d.onBind = fn
}

func (d *DownTrack) OnRTCP(fn func([]rtcp.Packet)) {
	d.onRTCP = fn
}

func (d *DownTrack) OnREMB(fn func(dt *DownTrack, remb *rtcp.ReceiverEstimatedMaximumBitrate)) {
	d.onREMB = fn
}

func (d *DownTrack) CurrentMaxLossFraction() uint8 {
	return d.lossFraction.get()
}

func (d *DownTrack) AddReceiverReportListener(listener ReceiverReportListener) {
	d.listenerLock.Lock()
	defer d.listenerLock.Unlock()
	d.receiverReportListeners = append(d.receiverReportListeners, listener)
}

func (d *DownTrack) OnAvailableLayersChanged(fn func(dt *DownTrack)) {
	d.onAvailableLayersChanged = fn
}

func (d *DownTrack) OnSubscriptionChanged(fn func(dt *DownTrack)) {
	d.onSubscriptionChanged = fn
}

func (d *DownTrack) OnSubscribedLayersChanged(fn func(dt *DownTrack, layers VideoLayers)) {
	d.onSubscribedLayersChanged = fn
}

func (d *DownTrack) OnPacketSent(fn func(dt *DownTrack, size int)) {
	d.onPacketSent = append(d.onPacketSent, fn)
}

func (d *DownTrack) Allocate(availableChannelCapacity int64) VideoAllocationResult {
	return d.forwarder.Allocate(availableChannelCapacity, d.receiver.GetBitrateTemporalCumulative())
}

func (d *DownTrack) TryAllocate(additionalChannelCapacity int64) VideoAllocationResult {
	return d.forwarder.TryAllocate(additionalChannelCapacity, d.receiver.GetBitrateTemporalCumulative())
}

func (d *DownTrack) FinalizeAllocate() {
	d.forwarder.FinalizeAllocate(d.receiver.GetBitrateTemporalCumulative())
}

func (d *DownTrack) AllocateNextHigher() bool {
	return d.forwarder.AllocateNextHigher(d.receiver.GetBitrateTemporalCumulative())
}

func (d *DownTrack) AllocationState() VideoAllocationState {
	return d.forwarder.AllocationState()
}

func (d *DownTrack) AllocationBandwidth() int64 {
	return d.forwarder.AllocationBandwidth()
}

func (d *DownTrack) CreateSourceDescriptionChunks() []rtcp.SourceDescriptionChunk {
	if !d.bound.get() {
		return nil
	}
	return []rtcp.SourceDescriptionChunk{
		{
			Source: d.ssrc,
			Items: []rtcp.SourceDescriptionItem{{
				Type: rtcp.SDESCNAME,
				Text: d.streamID,
			}},
		}, {
			Source: d.ssrc,
			Items: []rtcp.SourceDescriptionItem{{
				Type: rtcp.SDESType(15),
				Text: d.transceiver.Mid(),
			}},
		},
	}
}

func (d *DownTrack) CreateSenderReport() *rtcp.SenderReport {
	if !d.bound.get() {
		return nil
	}

	currentSpatialLayer := d.forwarder.CurrentSpatialLayer()
	if currentSpatialLayer == InvalidSpatialLayer {
		return nil
	}

	srRTP, srNTP := d.receiver.GetSenderReportTime(currentSpatialLayer)
	if srRTP == 0 {
		return nil
	}

	now := time.Now()
	nowNTP := toNtpTime(now)

	diff := (uint64(now.Sub(ntpTime(srNTP).Time())) * uint64(d.codec.ClockRate)) / uint64(time.Second)
	octets, packets := d.getSRStats()
	return &rtcp.SenderReport{
		SSRC:        d.ssrc,
		NTPTime:     uint64(nowNTP),
		RTPTime:     srRTP + uint32(diff),
		PacketCount: packets,
		OctetCount:  octets,
	}
}

func (d *DownTrack) UpdateStats(packetLen uint32) {
	d.octetCount.add(packetLen)
	d.packetCount.add(1)
}

func (d *DownTrack) writeBlankFrameRTP() error {
	// don't send if nothing has been sent
	if d.packetCount.get() == 0 {
		return nil
	}

	// LK-TODO: Support other video codecs
	if d.kind == webrtc.RTPCodecTypeAudio || (d.mime != "video/vp8" && d.mime != "video/h264") {
		return nil
	}

	snts, frameEndNeeded, err := d.forwarder.GetSnTsForBlankFrames()
	if err != nil {
		return err
	}

	// send a number of blank frames just in case there is loss.
	// Intentionally ignoring check for mute or bandwidth constrained mute
	// as this is used to clear client side buffer.
	for i := 0; i < len(snts); i++ {
		hdr := rtp.Header{
			Version:        2,
			Padding:        false,
			Marker:         true,
			PayloadType:    d.payloadType,
			SequenceNumber: snts[i].sequenceNumber,
			Timestamp:      snts[i].timestamp,
			SSRC:           d.ssrc,
			CSRC:           []uint32{},
		}

		err = d.writeRTPHeaderExtensions(&hdr)
		if err != nil {
			return err
		}

		switch d.mime {
		case "video/vp8":
			err = d.writeVP8BlankFrame(&hdr, frameEndNeeded)
		case "video/h264":
			err = d.writeH264BlankFrame(&hdr, frameEndNeeded)
		default:
			return nil
		}

		// only the first frame will need frameEndNeeded to close out the
		// previous picture, rest are small key frames
		frameEndNeeded = false
	}

	return nil
}

func (d *DownTrack) writeVP8BlankFrame(hdr *rtp.Header, frameEndNeeded bool) error {
	blankVP8 := d.forwarder.GetPaddingVP8(frameEndNeeded)

	// 1x1 key frame
	// Used even when closing out a previous frame. Looks like receivers
	// do not care about content (it will probably end up being an undecodable
	// frame, but that should be okay as there are key frames following)
	payload := make([]byte, blankVP8.HeaderSize+len(VP8KeyFrame1x1))
	vp8Header := payload[:blankVP8.HeaderSize]
	err := blankVP8.MarshalTo(vp8Header)
	if err != nil {
		return err
	}

	copy(payload[blankVP8.HeaderSize:], VP8KeyFrame1x1)

	_, err = d.writeStream.WriteRTP(hdr, payload)
	return err
}

func (d *DownTrack) writeH264BlankFrame(hdr *rtp.Header, frameEndNeeded bool) error {
	// TODO - Jie Zeng
	// now use STAP-A to compose sps, pps, idr together, most decoder support packetization-mode 1.
	// if client only support packetization-mode 0, use single nalu unit packet
	buf := make([]byte, 1462)
	offset := 0
	buf[0] = 0x18 // STAP-A
	offset++
	for _, payload := range H264KeyFrame2x2 {
		binary.BigEndian.PutUint16(buf[offset:], uint16(len(payload)))
		offset += 2
		copy(buf[offset:offset+len(payload)], payload)
		offset += len(payload)
	}
	_, err := d.writeStream.WriteRTP(hdr, buf[:offset])
	return err
}

func (d *DownTrack) handleRTCP(bytes []byte) {
	pkts, err := rtcp.Unmarshal(bytes)
	if err != nil {
		Logger.Error(err, "Unmarshal rtcp receiver packets err")
		return
	}

	if d.onRTCP != nil {
		d.onRTCP(pkts)
	}

	pliOnce := true
	sendPliOnce := func() {
		if pliOnce {
			targetSpatialLayer := d.forwarder.TargetSpatialLayer()
			if targetSpatialLayer != InvalidSpatialLayer {
				d.lastPli.set(time.Now().UnixNano())
				d.receiver.SendPLI(targetSpatialLayer)
				pliOnce = false
			}
		}
	}

	maxRatePacketLoss := uint8(0)
	for _, pkt := range pkts {
		switch p := pkt.(type) {
		case *rtcp.PictureLossIndication:
			sendPliOnce()
		case *rtcp.FullIntraRequest:
			sendPliOnce()
		case *rtcp.ReceiverEstimatedMaximumBitrate:
			if d.onREMB != nil {
				d.onREMB(d, p)
			}
		case *rtcp.ReceiverReport:
			// create new receiver report w/ only valid reception reports
			rr := &rtcp.ReceiverReport{
				SSRC:              p.SSRC,
				ProfileExtensions: p.ProfileExtensions,
			}
			for _, r := range p.Reports {
				if r.SSRC != d.ssrc {
					continue
				}
				rr.Reports = append(rr.Reports, r)
				if maxRatePacketLoss == 0 || maxRatePacketLoss < r.FractionLost {
					maxRatePacketLoss = r.FractionLost
				}
			}
			d.lossFraction.set(maxRatePacketLoss)
			if len(rr.Reports) > 0 {
				d.listenerLock.RLock()
				for _, l := range d.receiverReportListeners {
					l(d, rr)
				}
				d.listenerLock.RUnlock()
			}
		case *rtcp.TransportLayerNack:
			var nackedPackets []packetMeta
			for _, pair := range p.Nacks {
				nackedPackets = append(nackedPackets, d.sequencer.getSeqNoPairs(pair.PacketList())...)
			}
			go d.retransmitPackets(nackedPackets)
		}
	}
}

func (d *DownTrack) maybeTranslateVP8(pkt *rtp.Packet, meta packetMeta) error {
	if d.mime != "video/vp8" || len(pkt.Payload) == 0 {
		return nil
	}

	var incomingVP8 buffer.VP8
	if err := incomingVP8.Unmarshal(pkt.Payload); err != nil {
		return err
	}

	translatedVP8 := meta.unpackVP8()
	payload, err := d.translateVP8Packet(pkt, &incomingVP8, translatedVP8)
	if err != nil {
		return err
	}

	pkt.Payload = payload
	return nil
}

func (d *DownTrack) retransmitPackets(nackedPackets []packetMeta) {
	src := PacketFactory.Get().(*[]byte)
	defer PacketFactory.Put(src)
	for _, meta := range nackedPackets {
		pktBuff := *src
		n, err := d.receiver.ReadRTP(pktBuff, meta.layer, meta.sourceSeqNo)
		if err != nil {
			if err == io.EOF {
				break
			}
			continue
		}
		var pkt rtp.Packet
		if err = pkt.Unmarshal(pktBuff[:n]); err != nil {
			continue
		}
		pkt.Header.SequenceNumber = meta.targetSeqNo
		pkt.Header.Timestamp = meta.timestamp
		pkt.Header.SSRC = d.ssrc
		pkt.Header.PayloadType = d.payloadType

		err = d.maybeTranslateVP8(&pkt, meta)
		if err != nil {
			Logger.Error(err, "translating VP8 packet err")
			continue
		}

		err = d.writeRTPHeaderExtensions(&pkt.Header)
		if err != nil {
			Logger.Error(err, "writing rtp header extensions err")
			continue
		}

		if _, err = d.writeStream.WriteRTP(&pkt.Header, pkt.Payload); err != nil {
			Logger.Error(err, "Writing rtx packet err")
		} else {
			d.UpdateStats(uint32(n))
		}
	}
}

func (d *DownTrack) getSRStats() (octets, packets uint32) {
	return d.octetCount.get(), d.packetCount.get()
}

// writes RTP header extensions of track
func (d *DownTrack) writeRTPHeaderExtensions(hdr *rtp.Header) error {
	// clear out extensions that may have been in the forwarded header
	hdr.Extension = false
	hdr.ExtensionProfile = 0
	hdr.Extensions = []rtp.Extension{}

	for _, ext := range d.rtpHeaderExtensions {
		if ext.URI != sdp.ABSSendTimeURI {
			// supporting only abs-send-time
			continue
		}

		sendTime := rtp.NewAbsSendTimeExtension(time.Now())
		b, err := sendTime.Marshal()
		if err != nil {
			return err
		}

		err = hdr.SetExtension(uint8(ext.ID), b)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *DownTrack) getTranslatedRTPHeader(extPkt *buffer.ExtPacket, tpRTP *TranslationParamsRTP) (*rtp.Header, error) {
	hdr := extPkt.Packet.Header
	hdr.PayloadType = d.payloadType
	hdr.Timestamp = tpRTP.timestamp
	hdr.SequenceNumber = tpRTP.sequenceNumber
	hdr.SSRC = d.ssrc

	err := d.writeRTPHeaderExtensions(&hdr)
	if err != nil {
		return nil, err
	}

	return &hdr, nil
}

func (d *DownTrack) translateVP8Packet(pkt *rtp.Packet, incomingVP8 *buffer.VP8, translatedVP8 *buffer.VP8) (buf []byte, err error) {
	buf = *d.payload
	buf = buf[:len(pkt.Payload)+translatedVP8.HeaderSize-incomingVP8.HeaderSize]

	srcPayload := pkt.Payload[incomingVP8.HeaderSize:]
	dstPayload := buf[translatedVP8.HeaderSize:]
	copy(dstPayload, srcPayload)

	hdr := buf[:translatedVP8.HeaderSize]
	err = translatedVP8.MarshalTo(hdr)

	return
}

func (d *DownTrack) DebugInfo() map[string]interface{} {
	rtpMungerParams := d.forwarder.GetRTPMungerParams()
	stats := map[string]interface{}{
		"HighestIncomingSN": rtpMungerParams.highestIncomingSN,
		"LastSN":            rtpMungerParams.lastSN,
		"SNOffset":          rtpMungerParams.snOffset,
		"LastTS":            rtpMungerParams.lastTS,
		"TSOffset":          rtpMungerParams.tsOffset,
		"LastMarker":        rtpMungerParams.lastMarker,
		"LastRTP":           d.lastRTP.get(),
		"LastPli":           d.lastPli.get(),
		"PacketsDropped":    d.pktsDropped.get(),
	}

	senderReport := d.CreateSenderReport()
	if senderReport != nil {
		stats["NTPTime"] = senderReport.NTPTime
		stats["RTPTime"] = senderReport.RTPTime
		stats["PacketCount"] = senderReport.PacketCount
	}

	return map[string]interface{}{
		"PeerID":              d.peerID,
		"TrackID":             d.id,
		"StreamID":            d.streamID,
		"SSRC":                d.ssrc,
		"MimeType":            d.codec.MimeType,
		"Bound":               d.bound.get(),
		"Muted":               d.forwarder.Muted(),
		"CurrentSpatialLayer": d.forwarder.CurrentSpatialLayer,
		"Stats":               stats,
	}
}
