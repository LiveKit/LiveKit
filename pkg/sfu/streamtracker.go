package sfu

import (
	"sync"
	"time"

	"github.com/livekit/livekit-server/pkg/utils"
	"github.com/livekit/protocol/logger"
	"go.uber.org/atomic"
)

type StreamStatus int32

func (s StreamStatus) String() string {
	switch s {
	case StreamStatusStopped:
		return "stopped"
	case StreamStatusActive:
		return "active"
	default:
		return "unknown"
	}
}

const (
	StreamStatusStopped StreamStatus = 0
	StreamStatusActive  StreamStatus = 1
)

type StreamTrackerParams struct {
	// number of samples needed per cycle
	SamplesRequired uint32

	// number of cycles needed to be active
	CyclesRequired uint32

	CycleDuration time.Duration

	Logger logger.Logger
}

// StreamTracker keeps track of packet flow and ensures a particular up track is consistently producing
// It runs its own goroutine for detection, and fires OnStatusChanged callback
type StreamTracker struct {
	params StreamTrackerParams

	onStatusChanged func(status StreamStatus)

	paused         atomic.Bool
	countSinceLast atomic.Uint32 // number of packets received since last check
	generation     atomic.Uint32

	initMu      sync.Mutex
	initialized bool

	statusMu sync.RWMutex
	status   StreamStatus

	// only access within detectWorker
	cycleCount uint32

	// only access by the same goroutine as Observe
	lastSN uint16

	callbacksQueue *utils.OpsQueue

	isStopped atomic.Bool
}

func NewStreamTracker(params StreamTrackerParams) *StreamTracker {
	s := &StreamTracker{
		params:         params,
		status:         StreamStatusStopped,
		callbacksQueue: utils.NewOpsQueue(params.Logger),
	}
	return s
}

func (s *StreamTracker) OnStatusChanged(f func(status StreamStatus)) {
	s.onStatusChanged = f
}

func (s *StreamTracker) Status() StreamStatus {
	s.statusMu.RLock()
	defer s.statusMu.RUnlock()

	return s.status
}

func (s *StreamTracker) maybeSetStatus(status StreamStatus) {
	changed := false
	s.statusMu.Lock()
	if s.status != status {
		s.status = status
		changed = true
	}
	s.statusMu.Unlock()

	if changed && s.onStatusChanged != nil {
		s.callbacksQueue.Enqueue(func() {
			s.onStatusChanged(status)
		})
	}
}

func (s *StreamTracker) maybeSetActive() {
	s.maybeSetStatus(StreamStatusActive)
}

func (s *StreamTracker) maybeSetStopped() {
	s.maybeSetStatus(StreamStatusStopped)
}

func (s *StreamTracker) init() {
	s.maybeSetActive()

	go s.detectWorker(s.generation.Load())
}

func (s *StreamTracker) Start() {
	s.callbacksQueue.Start()
}

func (s *StreamTracker) Stop() {
	if s.isStopped.Swap(true) {
		return
	}

	s.callbacksQueue.Stop()

	// bump generation to trigger exit of worker
	s.generation.Inc()
}

func (s *StreamTracker) Reset() {
	if s.isStopped.Load() {
		return
	}

	// bump generation to trigger exit of current worker
	s.generation.Inc()

	s.countSinceLast.Store(0)
	s.cycleCount = 0

	s.initMu.Lock()
	s.initialized = false
	s.initMu.Unlock()

	s.statusMu.Lock()
	s.status = StreamStatusStopped
	s.statusMu.Unlock()
}

func (s *StreamTracker) SetPaused(paused bool) {
	s.paused.Store(paused)
}

// Observe a packet that's received
func (s *StreamTracker) Observe(sn uint16) {
	if s.paused.Load() {
		return
	}

	s.initMu.Lock()
	if !s.initialized {
		// first packet
		s.initialized = true
		s.initMu.Unlock()

		s.lastSN = sn
		s.countSinceLast.Inc()

		// declare stream active and start the detection worker
		go s.init()

		return
	}
	s.initMu.Unlock()

	// ignore out-of-order SNs
	if (sn - s.lastSN) > uint16(1<<15) {
		return
	}
	s.lastSN = sn
	s.countSinceLast.Inc()
}

func (s *StreamTracker) detectWorker(generation uint32) {
	ticker := time.NewTicker(s.params.CycleDuration)

	for {
		<-ticker.C
		if generation != s.generation.Load() {
			return
		}

		s.detectChanges()
	}
}

func (s *StreamTracker) detectChanges() {
	if s.paused.Load() {
		return
	}

	if s.countSinceLast.Load() >= s.params.SamplesRequired {
		s.cycleCount += 1
	} else {
		s.cycleCount = 0
	}

	if s.cycleCount == 0 {
		// flip to stopped
		s.maybeSetStopped()
	} else if s.cycleCount >= s.params.CyclesRequired {
		// flip to active
		s.maybeSetActive()
	}

	s.countSinceLast.Store(0)
}
