package sfu

import (
	"sync"
	"sync/atomic"
	"time"
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

// StreamTracker keeps track of packet flow and ensures a particular uptrack is consistently producing
// It runs its own goroutine for detection, and fires OnStatusChanged callback
type StreamTracker struct {
	// number of samples needed per cycle
	samplesRequired uint32
	// number of cycles needed to be active
	cyclesRequired uint64
	cycleDuration  time.Duration

	onStatusChanged func(status StreamStatus)

	paused         atomicBool
	countSinceLast uint32 // number of packets received since last check
	running        chan struct{}

	initMu      sync.Mutex
	initialized bool

	statusMu sync.RWMutex
	status   StreamStatus

	// only access within detectWorker
	cycleCount uint64

	// only access by the same goroutine as Observe
	lastSN uint16
}

func NewStreamTracker(samplesRequired uint32, cyclesRequired uint64, cycleDuration time.Duration) *StreamTracker {
	s := &StreamTracker{
		samplesRequired: samplesRequired,
		cyclesRequired:  cyclesRequired,
		cycleDuration:   cycleDuration,
		status:          StreamStatusStopped,
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

func (s *StreamTracker) maybeSetActive() {
	changed := false
	s.statusMu.Lock()
	if s.status != StreamStatusActive {
		s.status = StreamStatusActive
		changed = true
	}
	s.statusMu.Unlock()

	if changed && s.onStatusChanged != nil {
		s.onStatusChanged(StreamStatusActive)
	}
}

func (s *StreamTracker) maybeSetStopped() {
	changed := false
	s.statusMu.Lock()
	if s.status != StreamStatusStopped {
		s.status = StreamStatusStopped
		changed = true
	}
	s.statusMu.Unlock()

	if changed && s.onStatusChanged != nil {
		s.onStatusChanged(StreamStatusStopped)
	}
}

func (s *StreamTracker) init() {
	s.maybeSetActive()

	if s.isRunning() {
		return
	}
	s.running = make(chan struct{})
	go s.detectWorker()
}

func (s *StreamTracker) Start() {
}

func (s *StreamTracker) Stop() {
	if s.running != nil {
		close(s.running)
		s.running = nil
	}
}

func (s *StreamTracker) SetPaused(paused bool) {
	s.paused.set(paused)
}

func (s *StreamTracker) isRunning() bool {
	if s.running == nil {
		return false
	}
	select {
	case <-s.running:
		return false
	default:
		return true
	}
}

// Observe a packet that's received
func (s *StreamTracker) Observe(sn uint16) {
	if s.paused.get() {
		return
	}

	s.initMu.Lock()
	if !s.initialized {
		// first packet
		s.lastSN = sn
		s.initialized = true
		s.initMu.Unlock()

		// declare stream active and start the detect worker
		go s.init()

		return
	}
	s.initMu.Unlock()

	// ignore out-of-order SNs
	if (sn - s.lastSN) > uint16(1<<15) {
		return
	}
	s.lastSN = sn
	atomic.AddUint32(&s.countSinceLast, 1)
}

func (s *StreamTracker) detectWorker() {
	ticker := time.NewTicker(s.cycleDuration)

	for s.isRunning() {
		<-ticker.C
		if !s.isRunning() {
			return
		}

		s.detectChanges()
	}
}

func (s *StreamTracker) detectChanges() {
	if s.paused.get() {
		return
	}

	if atomic.LoadUint32(&s.countSinceLast) >= s.samplesRequired {
		s.cycleCount += 1
	} else {
		s.cycleCount = 0
	}

	if s.cycleCount == 0 {
		// flip to stopped
		s.maybeSetStopped()
	} else if s.cycleCount >= s.cyclesRequired {
		// flip to active
		s.maybeSetActive()
	}

	atomic.StoreUint32(&s.countSinceLast, 0)
}
