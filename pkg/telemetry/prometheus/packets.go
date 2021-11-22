package prometheus

import (
	"sync/atomic"

	"github.com/prometheus/client_golang/prometheus"
)

type Direction string

const (
	Incoming Direction = "incoming"
	Outgoing Direction = "outgoing"
)

var (
	atomicBytesIn    uint64
	atomicBytesOut   uint64
	atomicPacketsIn  uint64
	atomicPacketsOut uint64
	atomicNackTotal  uint64

	promPacketLabels = []string{"direction"}

	promPacketTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: livekitNamespace,
		Subsystem: "packet",
		Name:      "total",
	}, promPacketLabels)
	promPacketBytes = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: livekitNamespace,
		Subsystem: "packet",
		Name:      "bytes",
	}, promPacketLabels)
	promNackTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: livekitNamespace,
		Subsystem: "nack",
		Name:      "total",
	}, promPacketLabels)
	promPliTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: livekitNamespace,
		Subsystem: "pli",
		Name:      "total",
	}, promPacketLabels)
	promFirTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: livekitNamespace,
		Subsystem: "fir",
		Name:      "total",
	}, promPacketLabels)
)

func initPacketStats() {
	prometheus.MustRegister(promPacketTotal)
	prometheus.MustRegister(promPacketBytes)
	prometheus.MustRegister(promNackTotal)
	prometheus.MustRegister(promPliTotal)
	prometheus.MustRegister(promFirTotal)
}

func IncrementPackets(direction Direction, count uint64) {
	promPacketTotal.WithLabelValues(string(direction)).Add(float64(count))
	if direction == Incoming {
		atomic.AddUint64(&atomicPacketsIn, count)
	} else {
		atomic.AddUint64(&atomicPacketsOut, count)
	}
}

func IncrementBytes(direction Direction, count uint64) {
	promPacketBytes.WithLabelValues(string(direction)).Add(float64(count))
	if direction == Incoming {
		atomic.AddUint64(&atomicBytesIn, count)
	} else {
		atomic.AddUint64(&atomicBytesOut, count)
	}
}

func IncrementRTCP(direction Direction, nack, pli, fir int32) {
	if nack > 0 {
		promNackTotal.WithLabelValues(string(direction)).Add(float64(nack))
		atomic.AddUint64(&atomicNackTotal, uint64(nack))
	}
	if pli > 0 {
		promPliTotal.WithLabelValues(string(direction)).Add(float64(pli))
	}
	if fir > 0 {
		promFirTotal.WithLabelValues(string(direction)).Add(float64(fir))
	}
}
