// RAJA-TODO
// maintain max_pps and do some adjustment for pps reduction
// maintain max fps and do some adjustment for fps reduction
package connectionquality

import (
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/livekit/protocol/livekit"
	"github.com/livekit/protocol/logger"
)

const (
	MaxMOS = float32(4.5)

	maxScore  = float64(100.0)
	poorScore = float64(50.0)

	waitForQualityPoor = 25 * time.Second
	waitForQualityGood = 15 * time.Second

	unmuteTimeThreshold = float64(0.5)
)

// ------------------------------------------

type windowStat struct {
	startedAt       time.Time
	duration        time.Duration
	packetsExpected uint32
	packetsLost     uint32
	rttMax          uint32
	jitterMax       float64
}

func (w *windowStat) calculateScore(plw float64) float64 {
	effectiveDelay := (float64(w.rttMax) / 2.0) + ((w.jitterMax * 2.0) / 1000.0)
	delayEffect := effectiveDelay / 40.0
	if effectiveDelay > 160.0 {
		delayEffect = (effectiveDelay - 120.0) / 10.0
	}

	lossEffect := float64(0.0)
	if w.packetsExpected > 0 {
		lossEffect = float64(w.packetsLost) * 100.0 / float64(w.packetsExpected)
	}
	lossEffect *= plw

	return maxScore - delayEffect - lossEffect
}

func (w *windowStat) getStartTime() time.Time {
	return w.startedAt
}

func (w *windowStat) String() string {
	return fmt.Sprintf("start: %+v, dur: %+v, pe: %d, pl: %d, rtt: %d, jitter: %0.2f",
		w.startedAt,
		w.duration,
		w.packetsExpected,
		w.packetsLost,
		w.rttMax,
		w.jitterMax,
	)
}

// ------------------------------------------

type windowScore struct {
	stat  *windowStat
	score float64
}

func newWindowScore(stat *windowStat, plw float64) *windowScore {
	return &windowScore{
		stat:  stat,
		score: stat.calculateScore(plw),
	}
}

func newWindowScoreWithScore(stat *windowStat, score float64) *windowScore {
	return &windowScore{
		stat:  stat,
		score: score,
	}
}

func (w *windowScore) getScore() float64 {
	return w.score
}

func (w *windowScore) getStartTime() time.Time {
	return w.stat.getStartTime()
}

func (w *windowScore) String() string {
	return fmt.Sprintf("stat: {%+v}, score: %0.2f", w.stat, w.score)
}

// ------------------------------------------

type qualityScorerState int

const (
	qualityScorerStateStable qualityScorerState = iota
	qualityScorerStateRecovering
)

func (q qualityScorerState) String() string {
	switch q {
	case qualityScorerStateStable:
		return "STABLE"
	case qualityScorerStateRecovering:
		return "RECOVERING"
	default:
		return fmt.Sprintf("%d", int(q))
	}
}

// ------------------------------------------

type qualityScorerParams struct {
	PacketLossWeight float64
	Logger           logger.Logger
}

type qualityScorer struct {
	params qualityScorerParams

	lock         sync.RWMutex
	lastUpdateAt time.Time

	score   float64
	state   qualityScorerState
	windows []*windowScore

	mutedAt   time.Time
	unmutedAt time.Time

	maxPPS float64
}

func newQualityScorer(params qualityScorerParams) *qualityScorer {
	return &qualityScorer{
		params: params,
		score:  maxScore,
		state:  qualityScorerStateStable,
	}
}

func (q *qualityScorer) Start(at time.Time) {
	q.lock.Lock()
	defer q.lock.Unlock()

	q.lastUpdateAt = at
}

func (q *qualityScorer) UpdateMute(isMuted bool, at time.Time) {
	q.lock.Lock()
	defer q.lock.Unlock()

	if isMuted {
		q.mutedAt = at

		// stable when muted
		q.state = qualityScorerStateStable
		q.windows = q.windows[:0]
		q.score = maxScore
	} else {
		q.unmutedAt = at
	}
}

func (q *qualityScorer) Update(stat *windowStat, at time.Time) {
	q.lock.Lock()
	defer q.lock.Unlock()

	fmt.Printf("update at: %+v\n", at) // REMOVE
	// nothing to do when muted or not unmuted for long enough
	// NOTE: it is possible that unmute -> mute -> unmute transition happens in the
	//       same analysis window. On a transition to mute, state immediately moves
	//       to stable and quality EXCELLENT for responsiveness. On an unmute, the
	//       entire window data is considered (as long as enough time has passed since
	//       unmute) include the data before mute.
	if q.isMuted() || !q.isUnmutedEnough(at) {
		fmt.Printf("muted: %+v, unmuted: %+v, mutesAt: %+v, unmutedAt: %+v\n", q.isMuted(), q.isUnmutedEnough(at), q.mutedAt, q.unmutedAt) // REMOVE
		q.lastUpdateAt = at
		return
	}

	var ws *windowScore
	if stat == nil {
		ws = newWindowScoreWithScore(&windowStat{
			startedAt: q.lastUpdateAt,
			duration:  at.Sub(q.lastUpdateAt),
		}, poorScore)
	} else {
		if stat.packetsExpected == 0 {
			ws = newWindowScoreWithScore(stat, poorScore)
		} else {
			ws = newWindowScore(stat, q.getPacketLossWeight(stat))
		}
	}
	q.params.Logger.Infow("quality stat", "stat", stat, "window", ws) // REMOVE
	score := ws.getScore()
	cq := scoreToConnectionQuality(score)
	fmt.Printf("quality stat: %+v, window: %+v, score: %0.2f, cq: %+v\n\n\n", stat, ws, score, cq) // REMOVE

	q.lastUpdateAt = at

	// transition to start of recovering on any quality drop
	// WARNING NOTE: comparing protobuf enum values directly (livekit.ConnectionQuality)
	if scoreToConnectionQuality(q.score) > cq {
		q.windows = []*windowScore{ws}
		q.state = qualityScorerStateRecovering
		q.score = score
		return
	}

	// if stable and quality continues to be EXCELLENT, stay there.
	if q.state == qualityScorerStateStable && cq == livekit.ConnectionQuality_EXCELLENT {
		q.score = score
		return
	}

	// when recovering, look at a longer window
	q.windows = append(q.windows, ws)
	if !q.prune(at) {
		fmt.Printf("prune returning: len: %d, windows: %+v\n\n\n", len(q.windows), q.windows) // REMOVE
		// minimum recovery duration not satisfied, hold at current quality
		return
	}

	// take median of scores in a longer window to prevent quality reporting oscillations
	sort.Slice(q.windows, func(i, j int) bool { return q.windows[i].getScore() < q.windows[j].getScore() })
	mid := (len(q.windows)+1)/2 - 1
	q.score = q.windows[mid].getScore()
	fmt.Printf("prune passing: len: %d, windows: %+v, mid: %d, score: %0.2f\n\n\n", len(q.windows), q.windows, mid, q.score) // REMOVE
	if scoreToConnectionQuality(q.score) == livekit.ConnectionQuality_EXCELLENT {
		q.state = qualityScorerStateStable
		q.windows = q.windows[:0]
	}
}

func (q *qualityScorer) isMuted() bool {
	return !q.mutedAt.IsZero() && (q.unmutedAt.IsZero() || q.mutedAt.After(q.unmutedAt))
}

func (q *qualityScorer) isUnmutedEnough(at time.Time) bool {
	var sinceUnmute time.Duration
	if q.unmutedAt.IsZero() {
		sinceUnmute = at.Sub(q.lastUpdateAt)
	} else {
		sinceUnmute = at.Sub(q.unmutedAt)
	}

	sinceLastUpdate := at.Sub(q.lastUpdateAt)
	fmt.Printf("sinceUnmute: %+v, sinceLastUpdate: %+v, m: %+v, u: %+v, now: %+v\n", sinceUnmute, sinceLastUpdate, q.mutedAt, q.unmutedAt, at) // REMOVE

	return sinceUnmute.Seconds()/sinceLastUpdate.Seconds() > unmuteTimeThreshold
}

func (q *qualityScorer) getPacketLossWeight(stat *windowStat) float64 {
	if stat == nil {
		return q.params.PacketLossWeight
	}

	// packet loss is weighted by comparing against max packet rate seen.
	// this is to handle situations like DTX in audio and variable bit rate tracks like screen share.
	// and the effect of loss is not pronounced in those scenarios (audio silence, statis screen share).
	// for example, DTX typically uses only 5% of packets of full packet rate. at that rate,
	// packet loss weight is reduced to ~22% of configured weight (i. e. sqrt(0.05) * configured weight)
	pps := float64(stat.packetsExpected) / stat.duration.Seconds()
	if pps > q.maxPPS {
		q.maxPPS = pps
	}

	return math.Sqrt(pps/q.maxPPS) * q.params.PacketLossWeight
}

func (q *qualityScorer) prune(at time.Time) bool {
	cq := scoreToConnectionQuality(q.score)

	var wait time.Duration
	if cq == livekit.ConnectionQuality_POOR {
		wait = waitForQualityPoor
	} else {
		wait = waitForQualityGood
	}

	startThreshold := at.Add(-wait)
	sort.Slice(q.windows, func(i, j int) bool { return q.windows[i].getStartTime().Before(q.windows[j].getStartTime()) })
	for idx := 0; idx < len(q.windows); idx++ {
		w := q.windows[idx]
		if w.getStartTime().Before(startThreshold) {
			continue
		}

		q.windows = q.windows[idx:]
		break
	}
	fmt.Printf("pruned windows: %+v, wait: %+v, startThreshold: %+v, cq: %+v, score: %0.2f, at: %+v\n\n", q.windows, wait, startThreshold, cq, q.score, at) // REMOVE

	// find the oldest window of given quality and check if enough wait happened
	for idx := 0; idx < len(q.windows); idx++ {
		fmt.Printf("cq: %+v, idx: %d, idx_cq: %+v, start: %+v, wait: %+v, since: %+v\n", cq, idx, scoreToConnectionQuality(q.windows[idx].getScore()), q.windows[idx].getStartTime(), wait, at.Sub(q.windows[idx].getStartTime())) // REMOVE
		if cq == scoreToConnectionQuality(q.windows[idx].getScore()) {
			return at.Sub(q.windows[idx].getStartTime()) >= wait
		}
	}

	return false
}

func (q *qualityScorer) GetScoreAndQuality() (float32, livekit.ConnectionQuality) {
	q.lock.RLock()
	defer q.lock.RUnlock()

	return float32(q.score), scoreToConnectionQuality(q.score)
}

func (q *qualityScorer) GetMOSAndQuality() (float32, livekit.ConnectionQuality) {
	q.lock.RLock()
	defer q.lock.RUnlock()

	return scoreToMOS(q.score), scoreToConnectionQuality(q.score)
}

// ------------------------------------------

func scoreToConnectionQuality(score float64) livekit.ConnectionQuality {
	if score > 80.0 {
		return livekit.ConnectionQuality_EXCELLENT
	}

	if score > 60.0 {
		return livekit.ConnectionQuality_GOOD
	}

	return livekit.ConnectionQuality_POOR
}

// ------------------------------------------

func scoreToMOS(score float64) float32 {
	if score <= 0.0 {
		return 1.0
	}

	if score >= 100.0 {
		return 4.5
	}

	return float32(1.0 + 0.035*score + (0.000007 * score * (score - 60.0) * (100.0 - score)))
}

// ------------------------------------------
