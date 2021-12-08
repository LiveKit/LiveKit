package prometheus

import (
	"sync/atomic"
	"time"

	livekit "github.com/livekit/protocol/livekit"
	"github.com/prometheus/client_golang/prometheus"
)

const livekitNamespace = "livekit"

var (
	MessageCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: livekitNamespace,
			Subsystem: "node",
			Name:      "messages",
		},
		[]string{"type", "status"},
	)

	ServiceOperationCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: livekitNamespace,
			Subsystem: "node",
			Name:      "service_operation",
		},
		[]string{"type", "status", "error_type"},
	)
)

func init() {
	prometheus.MustRegister(MessageCounter)
	prometheus.MustRegister(ServiceOperationCounter)

	initPacketStats()
	initRoomStats()
}

func UpdateCurrentNodeStats(nodeStats *livekit.NodeStats) error {
	updatedAtPrevious := nodeStats.UpdatedAt
	nodeStats.UpdatedAt = time.Now().Unix()
	secondsSinceLastUpdate := nodeStats.UpdatedAt - updatedAtPrevious

	err := updateCurrentNodeSystemStats(nodeStats)
	updateCurrentNodeRoomStats(nodeStats)
	updateCurrentNodePacketStats(nodeStats, secondsSinceLastUpdate)

	return err
}

func updateCurrentNodeRoomStats(nodeStats *livekit.NodeStats) {
	nodeStats.NumClients = atomic.LoadInt32(&atomicParticipantTotal)
	nodeStats.NumRooms = atomic.LoadInt32(&atomicRoomTotal)
	nodeStats.NumTracksIn = atomic.LoadInt32(&atomicTrackPublishedTotal)
	nodeStats.NumTracksOut = atomic.LoadInt32(&atomicTrackSubscribedTotal)
}

func updateCurrentNodePacketStats(nodeStats *livekit.NodeStats, secondsSinceLastUpdate int64) {
	if secondsSinceLastUpdate == 0 {
		return
	}

	bytesInPrevious := nodeStats.BytesIn
	bytesOutPrevious := nodeStats.BytesOut
	packetsInPrevious := nodeStats.PacketsIn
	packetsOutPrevious := nodeStats.PacketsOut
	nackTotalPrevious := nodeStats.NackTotal

	nodeStats.BytesIn = atomic.LoadUint64(&atomicBytesIn)
	nodeStats.BytesOut = atomic.LoadUint64(&atomicBytesOut)
	nodeStats.PacketsIn = atomic.LoadUint64(&atomicPacketsIn)
	nodeStats.PacketsOut = atomic.LoadUint64(&atomicPacketsOut)
	nodeStats.NackTotal = atomic.LoadUint64(&atomicNackTotal)

	nodeStats.BytesInPerSec = perSec(bytesInPrevious, nodeStats.BytesIn, secondsSinceLastUpdate)
	nodeStats.BytesOutPerSec = perSec(bytesOutPrevious, nodeStats.BytesOut, secondsSinceLastUpdate)
	nodeStats.PacketsInPerSec = perSec(packetsInPrevious, nodeStats.PacketsIn, secondsSinceLastUpdate)
	nodeStats.PacketsOutPerSec = perSec(packetsOutPrevious, nodeStats.PacketsOut, secondsSinceLastUpdate)
	nodeStats.NackPerSec = perSec(nackTotalPrevious, nodeStats.NackTotal, secondsSinceLastUpdate)
}

func perSec(prev, curr uint64, secs int64) float32 {
	return float32(curr-prev) / float32(secs)
}
