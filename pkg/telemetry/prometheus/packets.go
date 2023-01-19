package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/atomic"

	"github.com/livekit/protocol/livekit"
)

type Direction string

const (
	Incoming               Direction = "incoming"
	Outgoing               Direction = "outgoing"
	transmissionInitial              = "initial"
	transmissionRetransmit           = "retransmit"
)

var (
	bytesIn           atomic.Uint64
	bytesOut          atomic.Uint64
	packetsIn         atomic.Uint64
	packetsOut        atomic.Uint64
	nackTotal         atomic.Uint64
	retransmitBytes   atomic.Uint64
	retransmitPackets atomic.Uint64
	participantJoin   atomic.Uint64
	participantRTC    atomic.Uint64

	promPacketLabels    = []string{"direction", "transmission"}
	promPacketTotal     *prometheus.CounterVec
	promPacketBytes     *prometheus.CounterVec
	promRTCPLabels      = []string{"direction"}
	promNackTotal       *prometheus.CounterVec
	promPliTotal        *prometheus.CounterVec
	promFirTotal        *prometheus.CounterVec
	promPacketLoss      *prometheus.HistogramVec
	promJitter          *prometheus.HistogramVec
	promRTT             *prometheus.HistogramVec
	promParticipantJoin *prometheus.CounterVec
	promConnections     *prometheus.GaugeVec
)

func initPacketStats(nodeID string, nodeType livekit.NodeType) {
	promPacketTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace:   livekitNamespace,
		Subsystem:   "packet",
		Name:        "total",
		ConstLabels: prometheus.Labels{"node_id": nodeID, "node_type": nodeType.String()},
	}, promPacketLabels)
	promPacketBytes = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace:   livekitNamespace,
		Subsystem:   "packet",
		Name:        "bytes",
		ConstLabels: prometheus.Labels{"node_id": nodeID, "node_type": nodeType.String()},
	}, promPacketLabels)
	promNackTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace:   livekitNamespace,
		Subsystem:   "nack",
		Name:        "total",
		ConstLabels: prometheus.Labels{"node_id": nodeID, "node_type": nodeType.String()},
	}, promRTCPLabels)
	promPliTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace:   livekitNamespace,
		Subsystem:   "pli",
		Name:        "total",
		ConstLabels: prometheus.Labels{"node_id": nodeID, "node_type": nodeType.String()},
	}, promRTCPLabels)
	promFirTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace:   livekitNamespace,
		Subsystem:   "fir",
		Name:        "total",
		ConstLabels: prometheus.Labels{"node_id": nodeID, "node_type": nodeType.String()},
	}, promRTCPLabels)
	promPacketLoss = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace:   livekitNamespace,
		Subsystem:   "packet_loss",
		Name:        "percent",
		ConstLabels: prometheus.Labels{"node_id": nodeID, "node_type": nodeType.String()},
		Buckets:     []float64{0.01, 0.02, 0.04, 0.07, 0.1, 0.2, 0.4, 0.6, 0.8, 1},
	}, promRTCPLabels)
	promJitter = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace:   livekitNamespace,
		Subsystem:   "jitter",
		Name:        "us",
		ConstLabels: prometheus.Labels{"node_id": nodeID, "node_type": nodeType.String()},
		Buckets:     []float64{100, 300, 500, 1000, 1500, 2000, 3000, 5000, 10000, 60000},
	}, promRTCPLabels)
	promRTT = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace:   livekitNamespace,
		Subsystem:   "rtt",
		Name:        "ms",
		ConstLabels: prometheus.Labels{"node_id": nodeID, "node_type": nodeType.String()},
		Buckets:     []float64{50, 75, 100, 125, 150, 200, 250, 400, 1000, 10000},
	}, promRTCPLabels)
	promParticipantJoin = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace:   livekitNamespace,
		Subsystem:   "participant_join",
		Name:        "total",
		ConstLabels: prometheus.Labels{"node_id": nodeID, "node_type": nodeType.String()},
	}, []string{"state"})
	promConnections = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace:   livekitNamespace,
		Subsystem:   "connection",
		Name:        "total",
		ConstLabels: prometheus.Labels{"node_id": nodeID, "node_type": nodeType.String()},
	}, []string{"kind"})

	prometheus.MustRegister(promPacketTotal)
	prometheus.MustRegister(promPacketBytes)
	prometheus.MustRegister(promNackTotal)
	prometheus.MustRegister(promPliTotal)
	prometheus.MustRegister(promFirTotal)
	prometheus.MustRegister(promPacketLoss)
	prometheus.MustRegister(promJitter)
	prometheus.MustRegister(promRTT)
	prometheus.MustRegister(promParticipantJoin)
	prometheus.MustRegister(promConnections)
}

func IncrementPackets(direction Direction, count uint64, retransmit bool) {
	promPacketTotal.WithLabelValues(
		string(direction),
		transmissionLabel(retransmit),
	).Add(float64(count))
	if direction == Incoming {
		packetsIn.Add(count)
	} else {
		packetsOut.Add(count)
		if retransmit {
			retransmitPackets.Add(count)
		}
	}
}

func IncrementBytes(direction Direction, count uint64, retransmit bool) {
	promPacketBytes.WithLabelValues(
		string(direction),
		transmissionLabel(retransmit),
	).Add(float64(count))
	if direction == Incoming {
		bytesIn.Add(count)
	} else {
		bytesOut.Add(count)
		if retransmit {
			retransmitBytes.Add(count)
		}
	}
}

func IncrementRTCP(direction Direction, nack, pli, fir uint32) {
	if nack > 0 {
		promNackTotal.WithLabelValues(string(direction)).Add(float64(nack))
		nackTotal.Add(uint64(nack))
	}
	if pli > 0 {
		promPliTotal.WithLabelValues(string(direction)).Add(float64(pli))
	}
	if fir > 0 {
		promFirTotal.WithLabelValues(string(direction)).Add(float64(fir))
	}
}

func RecordPacketLoss(direction Direction, loss float64) {
	if loss > 0 {
		promJitter.WithLabelValues(string(direction)).Observe(loss)
	}
}

func RecordJitter(direction Direction, jitter uint32) {
	if jitter > 0 {
		promJitter.WithLabelValues(string(direction)).Observe(float64(jitter))
	}
}

func RecordRTT(direction Direction, rtt uint32) {
	if rtt > 0 {
		promRTT.WithLabelValues(string(direction)).Observe(float64(rtt))
	}
}

func IncrementParticipantJoin(join uint32, rtcConnected ...bool) {
	if join > 0 {
		if len(rtcConnected) > 0 && rtcConnected[0] {
			participantRTC.Add(uint64(join))
			promParticipantJoin.WithLabelValues("rtc_connected").Add(float64(join))
		} else {
			participantJoin.Add(uint64(join))
			promParticipantJoin.WithLabelValues("signal_connected").Add(float64(join))
		}
	}
}

func AddConnection(direction Direction) {
	promConnections.WithLabelValues(string(direction)).Add(1)
}

func SubConnection(direction Direction) {
	promConnections.WithLabelValues(string(direction)).Sub(1)
}

func transmissionLabel(retransmit bool) string {
	if !retransmit {
		return transmissionInitial
	} else {
		return transmissionRetransmit
	}
}
