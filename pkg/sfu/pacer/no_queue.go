package pacer

import (
	"sync"

	"github.com/frostbyte73/core"
	"github.com/gammazero/deque"
	"github.com/livekit/livekit-server/pkg/sfu/sendsidebwe"
	"github.com/livekit/protocol/logger"
)

type NoQueue struct {
	*Base

	logger logger.Logger

	lock    sync.RWMutex
	packets deque.Deque[Packet]
	wake    chan struct{}
	stop    core.Fuse
}

func NewNoQueue(logger logger.Logger, sendSideBWE *sendsidebwe.SendSideBWE) *NoQueue {
	n := &NoQueue{
		Base:   NewBase(logger, sendSideBWE),
		logger: logger,
		wake:   make(chan struct{}, 1),
		stop:   core.NewFuse(),
	}
	n.packets.SetMinCapacity(9)

	go n.sendWorker()
	return n
}

func (n *NoQueue) Stop() {
	n.stop.Once(func() {
		close(n.wake)
	})
}

func (n *NoQueue) Enqueue(p Packet) {
	n.lock.Lock()
	n.packets.PushBack(p)

	notify := false
	if n.packets.Len() == 1 {
		notify = true
	}
	n.lock.Unlock()

	if notify {
		select {
		case n.wake <- struct{}{}:
		default:
		}
	}
}

func (n *NoQueue) sendWorker() {
	for {
		select {
		case <-n.wake:
			for {
				n.lock.Lock()
				if n.packets.Len() == 0 {
					n.lock.Unlock()
					break
				}
				p := n.packets.PopFront()
				n.lock.Unlock()

				n.Base.SendPacket(&p)
			}

		case <-n.stop.Watch():
			return
		}
	}
}

// ------------------------------------------------
