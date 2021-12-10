package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/livekit/protocol/logger"
	livekit "github.com/livekit/protocol/livekit"

	"github.com/livekit/livekit-server/pkg/config"
	"github.com/livekit/livekit-server/pkg/routing"
	"github.com/livekit/livekit-server/pkg/rtc"
	"github.com/livekit/livekit-server/pkg/rtc/types"
	"github.com/livekit/livekit-server/pkg/telemetry"
)

const (
	roomPurgeSeconds = 24 * 60 * 60
)

// RoomManager manages rooms and its interaction with participants.
// It's responsible for creating, deleting rooms, as well as running sessions for participants
type RoomManager struct {
	lock sync.RWMutex

	config      *config.Config
	rtcConfig   *rtc.WebRTCConfig
	currentNode routing.LocalNode
	router      routing.Router
	roomStore   RoomStore
	telemetry   telemetry.TelemetryService

	rooms map[string]*rtc.Room
}

func NewLocalRoomManager(
	conf *config.Config,
	roomStore RoomStore,
	currentNode routing.LocalNode,
	router routing.Router,
	telemetry telemetry.TelemetryService,
) (*RoomManager, error) {

	rtcConf, err := rtc.NewWebRTCConfig(conf, currentNode.Ip)
	if err != nil {
		return nil, err
	}

	r := &RoomManager{
		config:      conf,
		rtcConfig:   rtcConf,
		currentNode: currentNode,
		router:      router,
		roomStore:   roomStore,
		telemetry:   telemetry,

		rooms: make(map[string]*rtc.Room),
	}

	// hook up to router
	router.OnNewParticipantRTC(r.StartSession)
	router.OnRTCMessage(r.handleRTCMessage)
	return r, nil
}

func (r *RoomManager) GetRoom(ctx context.Context, roomName string) *rtc.Room {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return r.rooms[roomName]
}

// DeleteRoom completely deletes all room information, including active sessions, room store, and routing info
func (r *RoomManager) DeleteRoom(ctx context.Context, roomName string) error {
	logger.Infow("deleting room state", "room", roomName)
	r.lock.Lock()
	delete(r.rooms, roomName)
	r.lock.Unlock()

	var err, err2 error
	wg := sync.WaitGroup{}
	wg.Add(2)
	// clear routing information
	go func() {
		defer wg.Done()
		err = r.router.ClearRoomState(ctx, roomName)
	}()
	// also delete room from db
	go func() {
		defer wg.Done()
		err2 = r.roomStore.DeleteRoom(ctx, roomName)
	}()

	wg.Wait()
	if err2 != nil {
		err = err2
	}

	return err
}

// CleanupRooms cleans up after old rooms that have been around for awhile
func (r *RoomManager) CleanupRooms() error {
	// cleanup rooms that have been left for over a day
	ctx := context.Background()
	rooms, err := r.roomStore.ListRooms(ctx)
	if err != nil {
		return err
	}

	now := time.Now().Unix()
	for _, room := range rooms {
		if (now - room.CreationTime) > roomPurgeSeconds {
			if err := r.DeleteRoom(ctx, room.Name); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *RoomManager) CloseIdleRooms() {
	r.lock.RLock()
	rooms := make([]*rtc.Room, 0, len(r.rooms))
	for _, rm := range r.rooms {
		rooms = append(rooms, rm)
	}
	r.lock.RUnlock()

	for _, room := range rooms {
		room.CloseIfEmpty()
	}
}

func (r *RoomManager) HasParticipants() bool {
	r.lock.RLock()
	defer r.lock.RUnlock()

	for _, room := range r.rooms {
		if len(room.GetParticipants()) != 0 {
			return true
		}
	}
	return false
}

func (r *RoomManager) Stop() {
	// disconnect all clients
	r.lock.RLock()
	rooms := make([]*rtc.Room, 0, len(r.rooms))
	for _, rm := range r.rooms {
		rooms = append(rooms, rm)
	}
	r.lock.RUnlock()

	for _, room := range rooms {
		for _, p := range room.GetParticipants() {
			_ = p.Close()
		}
		room.Close()
	}

	if r.rtcConfig != nil {
		if r.rtcConfig.UDPMuxConn != nil {
			_ = r.rtcConfig.UDPMuxConn.Close()
		}
		if r.rtcConfig.TCPMuxListener != nil {
			_ = r.rtcConfig.TCPMuxListener.Close()
		}
	}
}

// StartSession starts WebRTC session when a new participant is connected, takes place on RTC node
func (r *RoomManager) StartSession(ctx context.Context, roomName string, pi routing.ParticipantInit, requestSource routing.MessageSource, responseSink routing.MessageSink) {
	room, err := r.getOrCreateRoom(ctx, roomName)
	if err != nil {
		logger.Errorw("could not create room", err, "room", roomName)
		return
	}

	participant := room.GetParticipant(pi.Identity)
	if participant != nil {
		// When reconnecting, it means WS has interrupted by underlying peer connection is still ok
		// in this mode, we'll keep the participant SID, and just swap the sink for the underlying connection
		if pi.Reconnect {
			logger.Debugw("resuming RTC session",
				"room", roomName,
				"nodeID", r.currentNode.Id,
				"participant", pi.Identity,
			)
			if err = room.ResumeParticipant(participant, responseSink); err != nil {
				logger.Warnw("could not resume participant", err,
					"participant", pi.Identity)
			}
			return
		} else {
			// we need to clean up the existing participant, so a new one can join
			room.RemoveParticipant(participant.Identity())
		}
	} else if pi.Reconnect {
		// send leave request if participant is trying to reconnect but missing from the room
		if err = responseSink.WriteMessage(&livekit.SignalResponse{
			Message: &livekit.SignalResponse_Leave{
				Leave: &livekit.LeaveRequest{
					CanReconnect: true,
				},
			},
		}); err != nil {
			logger.Warnw("could not restart participant", err,
				"participant", pi.Identity)
		}
		return
	}

	logger.Debugw("starting RTC session",
		"room", roomName,
		"nodeID", r.currentNode.Id,
		"participant", pi.Identity,
		"sdk", pi.Client.Sdk,
		"sdkVersion", pi.Client.Version,
		"protocol", pi.Client.Protocol,
	)

	pv := types.ProtocolVersion(pi.Client.Protocol)
	rtcConf := *r.rtcConfig
	rtcConf.SetBufferFactory(room.GetBufferFactor())
	participant, err = rtc.NewParticipant(rtc.ParticipantParams{
		Identity:        pi.Identity,
		Config:          &rtcConf,
		Sink:            responseSink,
		AudioConfig:     r.config.Audio,
		ProtocolVersion: pv,
		Telemetry:       r.telemetry,
		ThrottleConfig:  r.config.RTC.PLIThrottle,
		EnabledCodecs:   room.Room.EnabledCodecs,
		Hidden:          pi.Hidden,
		Logger:          room.Logger,
	})
	if err != nil {
		logger.Errorw("could not create participant", err)
		return
	}
	if pi.Metadata != "" {
		participant.SetMetadata(pi.Metadata)
	}

	if pi.Permission != nil {
		participant.SetPermission(pi.Permission)
	}

	// join room
	opts := rtc.ParticipantOptions{
		AutoSubscribe: pi.AutoSubscribe,
	}
	if err = room.Join(participant, &opts, r.iceServersForRoom(room.Room)); err != nil {
		logger.Errorw("could not join room", err)
		return
	}
	if err = r.roomStore.StoreParticipant(ctx, roomName, participant.ToProto()); err != nil {
		logger.Errorw("could not store participant", err)
	}
	// update roomstore with new numParticipants
	if !participant.Hidden() {
		err = r.roomStore.StoreRoom(ctx, room.Room)
		if err != nil {
			logger.Errorw("could not store room", err)
		}
	}

	r.telemetry.ParticipantJoined(ctx, room.Room, participant.ToProto())
	participant.OnClose(func(p types.Participant) {
		if err := r.roomStore.DeleteParticipant(ctx, roomName, p.Identity()); err != nil {
			logger.Errorw("could not delete participant", err)
		}
		// update roomstore with new numParticipants
		if !participant.Hidden() {
			err = r.roomStore.StoreRoom(ctx, room.Room)
			if err != nil {
				logger.Errorw("could not store room", err)
			}
		}
		r.telemetry.ParticipantLeft(ctx, room.Room, p.ToProto())
	})

	go r.rtcSessionWorker(room, participant, requestSource)
}

// create the actual room object, to be used on RTC node
func (r *RoomManager) getOrCreateRoom(ctx context.Context, roomName string) (*rtc.Room, error) {
	r.lock.RLock()
	room := r.rooms[roomName]
	r.lock.RUnlock()

	if room != nil {
		return room, nil
	}

	// create new room, get details first
	ri, err := r.roomStore.LoadRoom(ctx, roomName)
	if err != nil {
		return nil, err
	}

	// construct ice servers
	room = rtc.NewRoom(ri, *r.rtcConfig, &r.config.Audio, r.telemetry)
	r.telemetry.RoomStarted(ctx, room.Room)

	room.OnClose(func() {
		r.telemetry.RoomEnded(ctx, room.Room)
		if err := r.DeleteRoom(ctx, roomName); err != nil {
			logger.Errorw("could not delete room", err)
		}

		logger.Infow("room closed")
	})
	room.OnMetadataUpdate(func(metadata string) {
		if err := r.roomStore.StoreRoom(ctx, room.Room); err != nil {
			logger.Errorw("could not handle metadata update", err)
		}
	})
	room.OnParticipantChanged(func(p types.Participant) {
		if p.State() != livekit.ParticipantInfo_DISCONNECTED {
			if err := r.roomStore.StoreParticipant(ctx, roomName, p.ToProto()); err != nil {
				logger.Errorw("could not handle participant change", err)
			}
		}
	})
	r.lock.Lock()
	r.rooms[roomName] = room
	r.lock.Unlock()

	return room, nil
}

// manages an RTC session for a participant, runs on the RTC node
func (r *RoomManager) rtcSessionWorker(room *rtc.Room, participant types.Participant, requestSource routing.MessageSource) {
	defer func() {
		logger.Debugw("RTC session finishing",
			"participant", participant.Identity(),
			"pID", participant.ID(),
			"room", room.Room.Name,
			"roomID", room.Room.Sid,
		)
		_ = participant.Close()
	}()
	defer rtc.Recover()

	for {
		select {
		case <-time.After(time.Millisecond * 50):
			// periodic check to ensure participant didn't become disconnected
			if participant.State() == livekit.ParticipantInfo_DISCONNECTED {
				return
			}
		case obj := <-requestSource.ReadChan():
			// In single node mode, the request source is directly tied to the signal message channel
			// this means ICE restart isn't possible in single node mode
			if obj == nil {
				return
			}

			req := obj.(*livekit.SignalRequest)
			if err := r.handleSignalRequest(room, participant, req); err != nil {
				// more specific errors are already logged
				// treat errors returned as fatal
				return
			}
		}
	}
}

func (r *RoomManager) handleSignalRequest(room *rtc.Room, participant types.Participant, req *livekit.SignalRequest) error {
	switch msg := req.Message.(type) {
	case *livekit.SignalRequest_Offer:
		_, err := participant.HandleOffer(rtc.FromProtoSessionDescription(msg.Offer))
		if err != nil {
			logger.Errorw("could not handle offer", err,
				"room", room.Room.Name,
				"participant", participant.Identity(),
				"pID", participant.ID(),
			)
			return err
		}
	case *livekit.SignalRequest_AddTrack:
		logger.Debugw("add track request",
			"room", room.Room.Name,
			"participant", participant.Identity(),
			"pID", participant.ID(),
			"track", msg.AddTrack.Cid)
		participant.AddTrack(msg.AddTrack)
	case *livekit.SignalRequest_Answer:
		sd := rtc.FromProtoSessionDescription(msg.Answer)
		if err := participant.HandleAnswer(sd); err != nil {
			logger.Errorw("could not handle answer", err,
				"room", room.Room.Name,
				"participant", participant.Identity(),
				"pID", participant.ID(),
			)
			// connection cannot be successful if we can't answer
			return err
		}
	case *livekit.SignalRequest_Trickle:
		candidateInit, err := rtc.FromProtoTrickle(msg.Trickle)
		if err != nil {
			logger.Warnw("could not decode trickle", err,
				"room", room.Room.Name,
				"participant", participant.Identity(),
				"pID", participant.ID(),
			)
			return nil
		}
		// logger.Debugw("adding peer candidate", "participant", participant.Identity())
		if err := participant.AddICECandidate(candidateInit, msg.Trickle.Target); err != nil {
			logger.Warnw("could not handle trickle", err,
				"room", room.Room.Name,
				"participant", participant.Identity(),
				"pID", participant.ID(),
			)
		}
	case *livekit.SignalRequest_Mute:
		participant.SetTrackMuted(msg.Mute.Sid, msg.Mute.Muted, false)
	case *livekit.SignalRequest_Subscription:
		if err := room.UpdateSubscriptions(participant, msg.Subscription.TrackSids, msg.Subscription.Subscribe); err != nil {
			logger.Warnw("could not update subscription", err,
				"room", room.Room.Name,
				"participant", participant.Identity(),
				"pID", participant.ID(),
				"tracks", msg.Subscription.TrackSids,
				"subscribe", msg.Subscription.Subscribe)
		}
	case *livekit.SignalRequest_TrackSetting:
		for _, sid := range msg.TrackSetting.TrackSids {
			subTrack := participant.GetSubscribedTrack(sid)
			if subTrack == nil {
				logger.Warnw("unable to find SubscribedTrack", nil,
					"room", room.Room.Name,
					"participant", participant.Identity(),
					"pID", participant.ID(),
					"track", sid)
				continue
			}

			// find the source PublishedTrack
			publisher := room.GetParticipant(subTrack.PublisherIdentity())
			if publisher == nil {
				logger.Warnw("unable to find publisher of SubscribedTrack", nil,
					"room", room.Room.Name,
					"participant", participant.Identity(),
					"pID", participant.ID(),
					"publisher", subTrack.PublisherIdentity(),
					"track", sid)
				continue
			}

			pubTrack := publisher.GetPublishedTrack(sid)
			if pubTrack == nil {
				logger.Warnw("unable to find PublishedTrack", nil,
					"room", room.Room.Name,
					"participant", publisher.Identity(),
					"pID", publisher.ID(),
					"track", sid)
				continue
			}

			// find quality for published track
			logger.Debugw("updating track settings",
				"room", room.Room.Name,
				"participant", participant.Identity(),
				"pID", participant.ID(),
				"settings", msg.TrackSetting)
			subTrack.UpdateSubscriberSettings(msg.TrackSetting)
		}
	case *livekit.SignalRequest_UpdateLayers:
		track := participant.GetPublishedTrack(msg.UpdateLayers.TrackSid)
		if track == nil {
			logger.Warnw("could not find published track", nil,
				"track", msg.UpdateLayers.TrackSid)
			return nil
		}
		track.UpdateVideoLayers(msg.UpdateLayers.Layers)
	case *livekit.SignalRequest_Leave:
		_ = participant.Close()
	}
	return nil
}

// handles RTC messages resulted from Room API calls
func (r *RoomManager) handleRTCMessage(ctx context.Context, roomName, identity string, msg *livekit.RTCNodeMessage) {
	r.lock.RLock()
	room := r.rooms[roomName]
	r.lock.RUnlock()

	if room == nil {
		logger.Warnw("Could not find room", nil, "room", roomName)
		return
	}

	participant := room.GetParticipant(identity)

	switch rm := msg.Message.(type) {
	case *livekit.RTCNodeMessage_RemoveParticipant:
		if participant == nil {
			return
		}
		logger.Infow("removing participant", "room", roomName, "participant", identity)
		room.RemoveParticipant(identity)
	case *livekit.RTCNodeMessage_MuteTrack:
		if participant == nil {
			return
		}
		logger.Debugw("setting track muted", "room", roomName, "participant", identity,
			"track", rm.MuteTrack.TrackSid, "muted", rm.MuteTrack.Muted)
		if !rm.MuteTrack.Muted && !r.config.Room.EnableRemoteUnmute {
			logger.Errorw("cannot unmute track, remote unmute is disabled", nil)
			return
		}
		participant.SetTrackMuted(rm.MuteTrack.TrackSid, rm.MuteTrack.Muted, true)
	case *livekit.RTCNodeMessage_UpdateParticipant:
		if participant == nil {
			return
		}
		logger.Debugw("updating participant", "room", roomName, "participant", identity)
		if rm.UpdateParticipant.Metadata != "" {
			participant.SetMetadata(rm.UpdateParticipant.Metadata)
		}
		if rm.UpdateParticipant.Permission != nil {
			participant.SetPermission(rm.UpdateParticipant.Permission)
		}
	case *livekit.RTCNodeMessage_DeleteRoom:
		for _, p := range room.GetParticipants() {
			_ = p.Close()
		}
		room.Close()
	case *livekit.RTCNodeMessage_UpdateSubscriptions:
		if participant == nil {
			return
		}
		logger.Debugw("updating participant subscriptions", "room", roomName, "participant", identity)
		if err := room.UpdateSubscriptions(participant, rm.UpdateSubscriptions.TrackSids, rm.UpdateSubscriptions.Subscribe); err != nil {
			logger.Warnw("could not update subscription", err,
				"participant", participant.Identity(),
				"pID", participant.ID(),
				"tracks", rm.UpdateSubscriptions.TrackSids,
				"subscribe", rm.UpdateSubscriptions.Subscribe)
		}
	case *livekit.RTCNodeMessage_SendData:
		logger.Debugw("SendData", "message", rm)
		up := &livekit.UserPacket{
			Payload:         rm.SendData.Data,
			DestinationSids: rm.SendData.DestinationSids,
		}
		room.SendDataPacket(up, rm.SendData.Kind)
	case *livekit.RTCNodeMessage_UpdateRoomMetadata:
		logger.Debugw("updating room", "room", roomName)
		room.SetMetadata(rm.UpdateRoomMetadata.Metadata)
	}
}

func (r *RoomManager) iceServersForRoom(ri *livekit.Room) []*livekit.ICEServer {
	var iceServers []*livekit.ICEServer

	hasSTUN := false
	if r.config.TURN.Enabled {
		var urls []string
		if r.config.TURN.UDPPort > 0 {
			// UDP TURN is used as STUN
			hasSTUN = true
			urls = append(urls, fmt.Sprintf("turn:%s:%d?transport=udp", r.config.RTC.NodeIP, r.config.TURN.UDPPort))
		}
		if r.config.TURN.TLSPort > 0 {
			urls = append(urls, fmt.Sprintf("turns:%s:443?transport=tcp", r.config.TURN.Domain))
		}
		if len(urls) > 0 {
			iceServers = append(iceServers, &livekit.ICEServer{
				Urls:       urls,
				Username:   ri.Name,
				Credential: ri.TurnPassword,
			})
		}
	}

	if len(r.config.RTC.StunServers) > 0 {
		hasSTUN = true
		iceServers = append(iceServers, iceServerForStunServers(r.config.RTC.StunServers))
	}

	if !hasSTUN {
		iceServers = append(iceServers, iceServerForStunServers(config.DefaultStunServers))
	}
	return iceServers
}

func iceServerForStunServers(servers []string) *livekit.ICEServer {
	iceServer := &livekit.ICEServer{}
	for _, stunServer := range servers {
		iceServer.Urls = append(iceServer.Urls, fmt.Sprintf("stun:%s", stunServer))
	}
	return iceServer
}
