package routing

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"

	"github.com/livekit/livekit-server/pkg/logger"
	"github.com/livekit/livekit-server/pkg/utils"
	"github.com/livekit/livekit-server/proto/livekit"
)

const (
	// expire participant mappings after a day
	participantMappingTTL = 24 * time.Hour
	statsUpdateInterval   = 2 * time.Second
)

// RedisRouter uses Redis pub/sub to route signaling messages across different nodes
// It relies on the RTC node to be the primary driver of the participant connection.
// Because
type RedisRouter struct {
	LocalRouter
	rc        *redis.Client
	ctx       context.Context
	isStarted utils.AtomicFlag

	// map of connectionId => SignalNodeSink
	signalSinks map[string]*SignalNodeSink

	pubsub *redis.PubSub
	cancel func()
}

func NewRedisRouter(currentNode LocalNode, rc *redis.Client) *RedisRouter {
	rr := &RedisRouter{
		LocalRouter: *NewLocalRouter(currentNode),
		rc:          rc,
		signalSinks: make(map[string]*SignalNodeSink),
	}
	rr.ctx, rr.cancel = context.WithCancel(context.Background())
	return rr
}

func (r *RedisRouter) RegisterNode() error {
	data, err := proto.Marshal((*livekit.Node)(r.currentNode))
	if err != nil {
		return err
	}
	if err := r.rc.HSet(r.ctx, NodesKey, r.currentNode.Id, data).Err(); err != nil {
		return errors.Wrap(err, "could not register node")
	}
	return nil
}

func (r *RedisRouter) UnregisterNode() error {
	// could be called after Stop(), so we'd want to use an unrelated context
	return r.rc.HDel(context.Background(), NodesKey, r.currentNode.Id).Err()
}

func (r *RedisRouter) RemoveDeadNodes() error {
	nodes, err := r.ListNodes()
	if err != nil {
		return err
	}
	for _, n := range nodes {
		if !IsAvailable(n) {
			if err := r.rc.HDel(context.Background(), NodesKey, n.Id).Err(); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *RedisRouter) GetNodeForRoom(roomName string) (string, error) {
	val, err := r.rc.HGet(r.ctx, NodeRoomKey, roomName).Result()
	if err != nil {
		err = errors.Wrap(err, "could not get node for room")
	}
	return val, err
}

func (r *RedisRouter) SetNodeForRoom(roomName string, nodeId string) error {
	return r.rc.HSet(r.ctx, NodeRoomKey, roomName, nodeId).Err()
}

func (r *RedisRouter) ClearRoomState(roomName string) error {
	if err := r.rc.HDel(r.ctx, NodeRoomKey, roomName).Err(); err != nil {
		return errors.Wrap(err, "could not clear room state")
	}
	return nil
}

func (r *RedisRouter) GetNode(nodeId string) (*livekit.Node, error) {
	data, err := r.rc.HGet(r.ctx, NodesKey, nodeId).Result()
	if err != nil {
		return nil, err
	}
	n := livekit.Node{}
	if err = proto.Unmarshal([]byte(data), &n); err != nil {
		return nil, err
	}
	return &n, nil
}

func (r *RedisRouter) ListNodes() ([]*livekit.Node, error) {
	items, err := r.rc.HVals(r.ctx, NodesKey).Result()
	if err != nil {
		return nil, errors.Wrap(err, "could not list nodes")
	}
	nodes := make([]*livekit.Node, 0, len(items))
	for _, item := range items {
		n := livekit.Node{}
		if err := proto.Unmarshal([]byte(item), &n); err != nil {
			return nil, err
		}
		nodes = append(nodes, &n)
	}
	return nodes, nil
}

// signal connection sets up paths to the RTC node, and starts to route messages to that message queue
func (r *RedisRouter) StartParticipantSignal(roomName, identity, metadata string, reconnect bool) (connectionId string, reqSink MessageSink, resSource MessageSource, err error) {
	// find the node where the room is hosted at
	rtcNode, err := r.GetNodeForRoom(roomName)
	if err != nil {
		return
	}

	// create a new connection id
	connectionId = utils.NewGuid("CO_")
	pKey := participantKey(roomName, identity)

	// map signal & rtc nodes
	if err = r.setParticipantSignalNode(connectionId, r.currentNode.Id); err != nil {
		return
	}

	sink := NewRTCNodeSink(r.rc, rtcNode, pKey)

	// sends a message to start session
	err = sink.WriteMessage(&livekit.StartSession{
		RoomName: roomName,
		Identity: identity,
		Metadata: metadata,
		// connection id is to allow the RTC node to identify where to route the message back to
		ConnectionId: connectionId,
		Reconnect:    reconnect,
	})
	if err != nil {
		return
	}

	// index by connectionId, since there may be multiple connections for the participant
	resChan := r.getOrCreateMessageChannel(r.responseChannels, connectionId)
	return connectionId, sink, resChan, nil
}

func (r *RedisRouter) CreateRTCSink(roomName, identity string) (MessageSink, error) {
	pkey := participantKey(roomName, identity)
	rtcNode, err := r.getParticipantRTCNode(pkey)
	if err != nil {
		return nil, err
	}

	return NewRTCNodeSink(r.rc, rtcNode, pkey), nil
}

func (r *RedisRouter) startParticipantRTC(ss *livekit.StartSession, participantKey string) error {
	// find the node where the room is hosted at
	rtcNode, err := r.GetNodeForRoom(ss.RoomName)
	if err != nil {
		return err
	}

	if rtcNode != r.currentNode.Id {
		logger.Errorw("called participant on incorrect node",
			"rtcNode", rtcNode, "currentNode", r.currentNode.Id)
		return ErrIncorrectRTCNode
	}

	if err := r.setParticipantRTCNode(participantKey, rtcNode); err != nil {
		return err
	}

	// find signal node to send responses back
	signalNode, err := r.getParticipantSignalNode(ss.ConnectionId)
	if err != nil {
		return err
	}

	// treat it as a new participant connecting
	if r.onNewParticipant == nil {
		return ErrHandlerNotDefined
	}

	if !ss.Reconnect {
		// when it's not reconnecting, we do not want to re-use the same response sink
		// the previous rtc worker thread is still consuming off of it.
		// we'll want to sever the connection and switch to the new one
		r.lock.Lock()
		requestChan, ok := r.requestChannels[participantKey]
		r.lock.Unlock()
		if ok {
			requestChan.Close()
		}
	}

	reqChan := r.getOrCreateMessageChannel(r.requestChannels, participantKey)
	resSink := NewSignalNodeSink(r.rc, signalNode, ss.ConnectionId)
	r.onNewParticipant(
		ss.RoomName,
		ss.Identity,
		ss.Metadata,
		ss.Reconnect,
		reqChan,
		resSink,
	)
	return nil
}

func (r *RedisRouter) Start() error {
	if !r.isStarted.TrySet(true) {
		return nil
	}
	go r.statsWorker()
	go r.redisWorker()
	return nil
}

func (r *RedisRouter) Stop() {
	if !r.isStarted.TrySet(false) {
		return
	}
	logger.Debugw("stopping RedisRouter")
	r.pubsub.Close()
	r.cancel()
}

func (r *RedisRouter) setParticipantRTCNode(participantKey, nodeId string) error {
	err := r.rc.Set(r.ctx, participantRTCKey(participantKey), nodeId, participantMappingTTL).Err()
	if err != nil {
		err = errors.Wrap(err, "could not set rtc node")
	}
	return err
}

func (r *RedisRouter) setParticipantSignalNode(connectionId, nodeId string) error {
	if err := r.rc.Set(r.ctx, participantSignalKey(connectionId), nodeId, participantMappingTTL).Err(); err != nil {
		return errors.Wrap(err, "could not set signal node")
	}
	return nil
}

func (r *RedisRouter) getParticipantRTCNode(participantKey string) (string, error) {
	return r.rc.Get(r.ctx, participantRTCKey(participantKey)).Result()
}

func (r *RedisRouter) getParticipantSignalNode(connectionId string) (nodeId string, err error) {
	return r.rc.Get(r.ctx, participantSignalKey(connectionId)).Result()
}

// update node stats and cleanup
func (r *RedisRouter) statsWorker() {
	for r.ctx.Err() == nil {
		// update periodically seconds
		select {
		case <-time.After(statsUpdateInterval):
			r.currentNode.Stats.UpdatedAt = time.Now().Unix()
			if err := r.RegisterNode(); err != nil {
				logger.Errorw("could not update node", "error", err)
			}
		case <-r.ctx.Done():
			return
		}
	}
}

// worker that consumes redis messages intended for this node
func (r *RedisRouter) redisWorker() {
	defer func() {
		logger.Debugw("finishing redisWorker", "node", r.currentNode.Id)
	}()
	logger.Debugw("starting redisWorker", "node", r.currentNode.Id)

	sigChannel := signalNodeChannel(r.currentNode.Id)
	rtcChannel := rtcNodeChannel(r.currentNode.Id)
	r.pubsub = r.rc.Subscribe(r.ctx, sigChannel, rtcChannel)
	for msg := range r.pubsub.Channel() {
		if msg == nil {
			return
		}

		if msg.Channel == sigChannel {
			sm := livekit.SignalNodeMessage{}
			if err := proto.Unmarshal([]byte(msg.Payload), &sm); err != nil {
				logger.Errorw("could not unmarshal signal message on sigchan", "error", err)
				continue
			}
			if err := r.handleSignalMessage(&sm); err != nil {
				logger.Errorw("error processing signal message", "error", err)
				continue
			}
		} else if msg.Channel == rtcChannel {
			rm := livekit.RTCNodeMessage{}
			if err := proto.Unmarshal([]byte(msg.Payload), &rm); err != nil {
				logger.Errorw("could not unmarshal RTC message on rtcchan", "error", err)
				continue
			}
			if err := r.handleRTCMessage(&rm); err != nil {
				logger.Errorw("error processing RTC message", "error", err)
				continue
			}
		}
	}
}

func (r *RedisRouter) handleSignalMessage(sm *livekit.SignalNodeMessage) error {
	connectionId := sm.ConnectionId

	r.lock.Lock()
	resSink := r.responseChannels[connectionId]
	r.lock.Unlock()

	// if a client closed the channel, then sent more messages after that,
	if resSink == nil {
		return nil
	}

	switch rmb := sm.Message.(type) {
	case *livekit.SignalNodeMessage_Response:
		//logger.Debugw("forwarding signal message",
		//	"connectionId", connectionId,
		//	"type", fmt.Sprintf("%T", rmb.Response.Message))
		if err := resSink.WriteMessage(rmb.Response); err != nil {
			return err
		}

	case *livekit.SignalNodeMessage_EndSession:
		//logger.Debugw("received EndSession, closing signal connection",
		//	"connectionId", connectionId)
		resSink.Close()
	}
	return nil
}

func (r *RedisRouter) handleRTCMessage(rm *livekit.RTCNodeMessage) error {
	pKey := rm.ParticipantKey

	switch rmb := rm.Message.(type) {
	case *livekit.RTCNodeMessage_StartSession:
		// RTC session should start on this node
		if err := r.startParticipantRTC(rmb.StartSession, pKey); err != nil {
			return errors.Wrap(err, "could not start participant")
		}

	case *livekit.RTCNodeMessage_Request:
		r.lock.Lock()
		requestChan := r.requestChannels[pKey]
		r.lock.Unlock()
		if err := requestChan.WriteMessage(rmb.Request); err != nil {
			return err
		}

	default:
		// route it to handler
		if r.onRTCMessage != nil {
			roomName, identity, err := parseParticipantKey(pKey)
			if err != nil {
				return err
			}
			r.onRTCMessage(roomName, identity, rm)
		}
	}
	return nil
}
