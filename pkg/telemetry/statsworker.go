package telemetry

import (
	"context"

	"google.golang.org/protobuf/proto"

	"github.com/livekit/protocol/livekit"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Stat struct {
	Score                  float32
	Rtt                    uint32
	Jitter                 uint32
	TotalPrimaryPackets    uint32
	TotalPrimaryBytes      uint64
	TotalRetransmitPackets uint32
	TotalRetransmitBytes   uint64
	TotalPaddingPackets    uint32
	TotalPaddingBytes      uint64
	TotalPacketsLost       uint32
	TotalFrames            uint32
	TotalNacks             uint32
	TotalPlis              uint32
	TotalFirs              uint32
	VideoLayers            map[int32]*livekit.AnalyticsVideoLayer
	TotalBytes             uint64
	TotalPackets           uint32
}

func (stat *Stat) ToAnalyticsStats(layers *livekit.AnalyticsVideoLayer) *livekit.AnalyticsStat {
	stream := &livekit.AnalyticsStream{
		TotalPrimaryPackets:    stat.TotalPrimaryPackets,
		TotalPrimaryBytes:      stat.TotalPrimaryBytes,
		TotalRetransmitPackets: stat.TotalRetransmitPackets,
		TotalRetransmitBytes:   stat.TotalRetransmitBytes,
		TotalPaddingPackets:    stat.TotalPaddingPackets,
		TotalPaddingBytes:      stat.TotalPaddingBytes,
		TotalPacketsLost:       stat.TotalPacketsLost,
		TotalFrames:            stat.TotalFrames,
		Rtt:                    stat.Rtt,
		Jitter:                 stat.Jitter,
		TotalNacks:             stat.TotalNacks,
		TotalPlis:              stat.TotalPlis,
		TotalFirs:              stat.TotalFirs,
		VideoLayers:            []*livekit.AnalyticsVideoLayer{layers},
	}
	return &livekit.AnalyticsStat{Streams: []*livekit.AnalyticsStream{stream}, Score: stat.Score}
}

type Stats struct {
	// local stats context used for coalesce/delta calculations
	curStats  *Stat
	prevStats *Stat

	// current stats received per stream
	queue []*livekit.AnalyticsStat
}

// StatsWorker handles participant stats
type StatsWorker struct {
	ctx           context.Context
	t             TelemetryReporter
	roomID        livekit.RoomID
	roomName      livekit.RoomName
	participantID livekit.ParticipantID

	outgoingPerTrack map[livekit.TrackID]Stats
	incomingPerTrack map[livekit.TrackID]Stats
}

func newStatsWorker(
	ctx context.Context,
	t TelemetryReporter,
	roomID livekit.RoomID,
	roomName livekit.RoomName,
	participantID livekit.ParticipantID,
) *StatsWorker {
	s := &StatsWorker{
		ctx:           ctx,
		t:             t,
		roomID:        roomID,
		roomName:      roomName,
		participantID: participantID,

		outgoingPerTrack: make(map[livekit.TrackID]Stats),
		incomingPerTrack: make(map[livekit.TrackID]Stats),
	}
	return s
}

func (s *StatsWorker) appendOutgoing(trackID livekit.TrackID, stat *livekit.AnalyticsStat) {
	stats := s.outgoingPerTrack[trackID]
	stats.queue = append(stats.queue, stat)
	s.outgoingPerTrack[trackID] = stats
}

func (s *StatsWorker) appendIncoming(trackID livekit.TrackID, stat *livekit.AnalyticsStat) {
	stats := s.incomingPerTrack[trackID]
	stats.queue = append(stats.queue, stat)
	s.incomingPerTrack[trackID] = stats
}

func (s *StatsWorker) OnTrackStat(trackID livekit.TrackID, direction livekit.StreamType, stat *livekit.AnalyticsStat) {
	if direction == livekit.StreamType_DOWNSTREAM {
		s.appendOutgoing(trackID, stat)
	} else {
		s.appendIncoming(trackID, stat)
	}
}

func (s *StatsWorker) Update() {
	ts := timestamppb.Now()

	stats := make([]*livekit.AnalyticsStat, 0, len(s.incomingPerTrack)+len(s.outgoingPerTrack))
	stats = s.collectUpstreamStats(ts, stats)
	stats = s.collectDownstreamStats(ts, stats)
	if len(stats) > 0 {
		s.t.Report(s.ctx, stats)
	}
}

func (s *StatsWorker) collectDownstreamStats(ts *timestamppb.Timestamp, stats []*livekit.AnalyticsStat) []*livekit.AnalyticsStat {
	for trackID, analyticsStats := range s.outgoingPerTrack {
		analyticsStat := s.computeDeltaStats(&analyticsStats, ts, trackID, livekit.StreamType_DOWNSTREAM)
		if analyticsStat != nil {
			stats = append(stats, analyticsStat)
		}
		// clear the queue
		analyticsStats.queue = nil
		s.outgoingPerTrack[trackID] = analyticsStats
	}
	return stats
}

func (s *StatsWorker) collectUpstreamStats(ts *timestamppb.Timestamp, stats []*livekit.AnalyticsStat) []*livekit.AnalyticsStat {
	for trackID, analyticsStats := range s.incomingPerTrack {
		analyticsStat := s.computeDeltaStats(&analyticsStats, ts, trackID, livekit.StreamType_UPSTREAM)
		if analyticsStat != nil {
			stats = append(stats, analyticsStat)
		}
		// clear the queue
		analyticsStats.queue = nil
		s.incomingPerTrack[trackID] = analyticsStats
	}
	return stats
}
func (s *StatsWorker) computeDeltaStats(stats *Stats, ts *timestamppb.Timestamp, trackID livekit.TrackID, kind livekit.StreamType) *livekit.AnalyticsStat {
	// merge all streams stats of track
	stats.coalesce()
	// create deltaStats to send
	analyticsStat := stats.computeDeltaStats()
	// update stats for next interval
	if analyticsStat == nil {
		return nil
	}
	stats.update()
	s.patch(analyticsStat, ts, trackID, kind)

	return analyticsStat
}

func (s *StatsWorker) patch(
	analyticsStat *livekit.AnalyticsStat,
	ts *timestamppb.Timestamp,
	trackID livekit.TrackID,
	kind livekit.StreamType,
) {
	analyticsStat.TimeStamp = ts
	analyticsStat.TrackId = string(trackID)
	analyticsStat.Kind = kind
	analyticsStat.RoomId = string(s.roomID)
	analyticsStat.ParticipantId = string(s.participantID)
	analyticsStat.RoomName = string(s.roomName)
}

func (s *StatsWorker) Close() {
	s.Update()
}

// create a single stream and single video layer post aggregation
func (stats *Stats) update() {
	stats.prevStats = stats.curStats
	stats.curStats = nil
}

// create a single stream and single video layer post aggregation
func (stats *Stats) coalesce() {
	if len(stats.queue) == 0 {
		return
	}

	// average score of all available stats
	score := float32(0.0)
	for _, stat := range stats.queue {
		score += stat.Score
	}
	score = score / float32(len(stats.queue))

	// aggregate streams across all stats
	maxRTT := make(map[uint32]uint32)
	maxJitter := make(map[uint32]uint32)
	analyticsStreams := make(map[uint32]*livekit.AnalyticsStream)
	for _, stat := range stats.queue {
		//
		// For each stream (identified by SSRC) consolidate reports.
		// For cumulative stats, take the latest report.
		// For instantaneous stats, take maximum (or some other appropriate representation)
		//
		for _, stream := range stat.Streams {
			ssrc := stream.Ssrc
			analyticsStream := analyticsStreams[ssrc]
			if analyticsStream == nil {
				analyticsStreams[ssrc] = stream
				maxRTT[ssrc] = stream.Rtt
				maxJitter[ssrc] = stream.Jitter
				continue
			}

			if stream.TotalPrimaryPackets <= analyticsStream.TotalPrimaryPackets {
				// total count should be monotonically increasing
				continue
			}

			analyticsStreams[ssrc] = stream
			if stream.Rtt > maxRTT[ssrc] {
				maxRTT[ssrc] = stream.Rtt
			}

			if stream.Jitter > maxJitter[ssrc] {
				maxJitter[ssrc] = stream.Jitter
			}
		}
	}

	curStats := Stat{Score: score, VideoLayers: make(map[int32]*livekit.AnalyticsVideoLayer)}
	// find aggregates across streams
	for ssrc, analyticsStream := range analyticsStreams {
		rtt := maxRTT[ssrc]
		if rtt > curStats.Rtt {
			curStats.Rtt = rtt
		}

		jitter := maxJitter[ssrc]
		if jitter > curStats.Jitter {
			curStats.Jitter = jitter
		}

		curStats.TotalFrames += analyticsStream.TotalFrames
		curStats.TotalPrimaryPackets += analyticsStream.TotalPrimaryPackets
		curStats.TotalPrimaryBytes += analyticsStream.TotalPrimaryBytes
		curStats.TotalRetransmitPackets += analyticsStream.TotalRetransmitPackets
		curStats.TotalRetransmitBytes += analyticsStream.TotalRetransmitBytes
		curStats.TotalPaddingPackets += analyticsStream.TotalPaddingPackets
		curStats.TotalPaddingBytes += analyticsStream.TotalPaddingBytes
		curStats.TotalPacketsLost += analyticsStream.TotalPacketsLost
		curStats.TotalFrames += analyticsStream.TotalFrames
		curStats.TotalNacks += analyticsStream.TotalNacks
		curStats.TotalPlis += analyticsStream.TotalPlis
		curStats.TotalFirs += analyticsStream.TotalFirs
		// add/update new video VideoLayers data to current and sum up video layer bytes/packets
		for _, videoLayer := range analyticsStream.VideoLayers {
			curStats.VideoLayers[videoLayer.Layer] = proto.Clone(videoLayer).(*livekit.AnalyticsVideoLayer)
			curStats.TotalPackets += videoLayer.TotalPackets
			curStats.TotalBytes += videoLayer.TotalBytes
		}

	}

	// update currentStats
	stats.curStats = &curStats

	return
}

// find delta between  curStats and prevStats and prepare proto payload
func (stats *Stats) computeDeltaStats() *livekit.AnalyticsStat {

	if stats.curStats == nil {
		return nil
	}

	// Stats in both queue/prev contain consolidated single deltaStats
	cur := stats.curStats

	var maxLayer int32
	var maxTotalBytes uint64
	//create a map of VideoLayers - to pick max/best layer wrt current and prev
	curLayers := make(map[int32]*livekit.AnalyticsVideoLayer)
	for _, layer := range cur.VideoLayers {
		curLayers[layer.Layer] = layer
		// identify layer which sent max data - as VideoLayers can change in current interval
		if layer.TotalBytes > maxTotalBytes {
			maxTotalBytes = layer.TotalBytes
			maxLayer = layer.Layer
		}
	}

	// no previous stats, prepare stat
	if stats.prevStats == nil {
		return cur.ToAnalyticsStats(curLayers[maxLayer])
	}

	// we have prevStats, find delta between cur and prev
	prev := stats.prevStats
	deltaStats := Stat{}
	deltaStats.Rtt = cur.Rtt
	deltaStats.Jitter = cur.Jitter
	deltaStats.TotalPlis = cur.TotalPlis - prev.TotalPlis
	deltaStats.TotalFrames = cur.TotalFrames - prev.TotalFrames
	deltaStats.TotalNacks = cur.TotalNacks - prev.TotalNacks
	deltaStats.TotalFirs = cur.TotalFirs - prev.TotalFirs
	deltaStats.TotalPacketsLost = cur.TotalPacketsLost - prev.TotalPacketsLost
	deltaStats.TotalPrimaryPackets = cur.TotalPrimaryPackets - prev.TotalPrimaryPackets
	deltaStats.TotalRetransmitPackets = cur.TotalRetransmitPackets - prev.TotalRetransmitPackets
	deltaStats.TotalPaddingPackets = cur.TotalPaddingPackets - prev.TotalPaddingPackets
	deltaStats.TotalPaddingPackets = cur.TotalPaddingPackets - prev.TotalPaddingPackets
	deltaStats.TotalPrimaryBytes = cur.TotalPrimaryBytes - prev.TotalPrimaryBytes
	deltaStats.TotalPaddingBytes = cur.TotalPaddingBytes - prev.TotalPaddingBytes
	deltaStats.TotalRetransmitBytes = cur.TotalRetransmitBytes - prev.TotalRetransmitBytes

	videoLayer := &livekit.AnalyticsVideoLayer{}
	if len(cur.VideoLayers) > 0 {
		// find the prev max layer to calculate frame deltas only
		if prevMaxLayer, ok := prev.VideoLayers[maxLayer]; ok {
			videoLayer.Layer = maxLayer
			videoLayer.TotalFrames = curLayers[maxLayer].TotalFrames - prevMaxLayer.TotalFrames
		} else {
			videoLayer = curLayers[maxLayer]
		}
		// we accumulate bytes/packets across layers
		videoLayer.TotalBytes = cur.TotalBytes - prev.TotalBytes
		videoLayer.TotalPackets = cur.TotalPackets - prev.TotalPackets
	}
	// if no packets from any layers, return nil to send no stats
	if deltaStats.TotalPackets == 0 && deltaStats.TotalPrimaryPackets == 0 && deltaStats.TotalRetransmitPackets == 0 && deltaStats.TotalPaddingPackets == 0 {
		return nil
	}
	return deltaStats.ToAnalyticsStats(videoLayer)
}
