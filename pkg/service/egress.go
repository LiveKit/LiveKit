// Copyright 2023 LiveKit, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/twitchtv/twirp"

	"github.com/livekit/livekit-server/pkg/rtc"
	"github.com/livekit/protocol/egress"
	"github.com/livekit/protocol/livekit"
	"github.com/livekit/protocol/logger"
	"github.com/livekit/protocol/rpc"
	"github.com/livekit/protocol/utils"
)

type EgressService struct {
	client      rpc.EgressClient
	io          IOClient
	roomService livekit.RoomService
	store       ServiceStore
	launcher    rtc.EgressLauncher
}

type egressLauncher struct {
	client rpc.EgressClient
	io     IOClient
}

func NewEgressLauncher(client rpc.EgressClient, io IOClient) rtc.EgressLauncher {
	if client == nil {
		return nil
	}
	return &egressLauncher{
		client: client,
		io:     io,
	}
}

func NewEgressService(
	client rpc.EgressClient,
	store ServiceStore,
	io IOClient,
	rs livekit.RoomService,
	launcher rtc.EgressLauncher,
) *EgressService {
	return &EgressService{
		client:      client,
		store:       store,
		io:          io,
		roomService: rs,
		launcher:    launcher,
	}
}

func (s *EgressService) StartRoomCompositeEgress(ctx context.Context, req *livekit.RoomCompositeEgressRequest) (*livekit.EgressInfo, error) {
	fields := []interface{}{
		"room", req.RoomName,
		"baseUrl", req.CustomBaseUrl,
		"outputType", egress.GetOutputType(req),
	}
	defer func() {
		AppendLogFields(ctx, fields...)
	}()
	ei, err := s.startEgress(ctx, livekit.RoomName(req.RoomName), &rpc.StartEgressRequest{
		Request: &rpc.StartEgressRequest_RoomComposite{
			RoomComposite: req,
		},
	})
	if err != nil {
		return nil, err
	}
	fields = append(fields, "egressID", ei.EgressId)
	return ei, err
}

func (s *EgressService) StartWebEgress(ctx context.Context, req *livekit.WebEgressRequest) (*livekit.EgressInfo, error) {
	fields := []interface{}{
		"url", req.Url,
		"outputType", egress.GetOutputType(req),
	}
	defer func() {
		AppendLogFields(ctx, fields...)
	}()
	ei, err := s.startEgress(ctx, "", &rpc.StartEgressRequest{
		Request: &rpc.StartEgressRequest_Web{
			Web: req,
		},
	})
	if err != nil {
		return nil, err
	}
	fields = append(fields, "egressID", ei.EgressId)
	return ei, err
}

func (s *EgressService) StartParticipantEgress(ctx context.Context, req *livekit.ParticipantEgressRequest) (*livekit.EgressInfo, error) {
	fields := []interface{}{
		"room", req.RoomName,
		"identity", req.Identity,
		"outputType", egress.GetOutputType(req),
	}
	defer func() {
		AppendLogFields(ctx, fields...)
	}()
	ei, err := s.startEgress(ctx, livekit.RoomName(req.RoomName), &rpc.StartEgressRequest{
		Request: &rpc.StartEgressRequest_Participant{
			Participant: req,
		},
	})
	if err != nil {
		return nil, err
	}
	fields = append(fields, "egressID", ei.EgressId)
	return ei, err
}

func (s *EgressService) StartTrackCompositeEgress(ctx context.Context, req *livekit.TrackCompositeEgressRequest) (*livekit.EgressInfo, error) {
	fields := []interface{}{
		"room", req.RoomName,
		"audioTrackID", req.AudioTrackId,
		"videoTrackID", req.VideoTrackId,
		"outputType", egress.GetOutputType(req),
	}
	defer func() {
		AppendLogFields(ctx, fields...)
	}()
	ei, err := s.startEgress(ctx, livekit.RoomName(req.RoomName), &rpc.StartEgressRequest{
		Request: &rpc.StartEgressRequest_TrackComposite{
			TrackComposite: req,
		},
	})
	if err != nil {
		return nil, err
	}
	fields = append(fields, "egressID", ei.EgressId)
	return ei, err
}

func (s *EgressService) StartTrackEgress(ctx context.Context, req *livekit.TrackEgressRequest) (*livekit.EgressInfo, error) {
	fields := []interface{}{"room", req.RoomName, "trackID", req.TrackId}
	if t := reflect.TypeOf(req.Output); t != nil {
		fields = append(fields, "outputType", t.String())
	}
	defer func() {
		AppendLogFields(ctx, fields...)
	}()
	ei, err := s.startEgress(ctx, livekit.RoomName(req.RoomName), &rpc.StartEgressRequest{
		Request: &rpc.StartEgressRequest_Track{
			Track: req,
		},
	})
	if err != nil {
		return nil, err
	}
	fields = append(fields, "egressID", ei.EgressId)
	return ei, err
}

func (s *EgressService) startEgress(ctx context.Context, roomName livekit.RoomName, req *rpc.StartEgressRequest) (*livekit.EgressInfo, error) {
	if err := EnsureRecordPermission(ctx); err != nil {
		return nil, twirpAuthError(err)
	} else if s.launcher == nil {
		return nil, ErrEgressNotConnected
	}
	if roomName != "" {
		room, _, err := s.store.LoadRoom(ctx, roomName, false)
		if err != nil {
			return nil, err
		}
		req.RoomId = room.Sid
	}
	return s.launcher.StartEgress(ctx, req)
}

func (s *egressLauncher) StartEgress(ctx context.Context, req *rpc.StartEgressRequest) (*livekit.EgressInfo, error) {
	info, err := s.StartEgressWithClusterId(ctx, "", req)
	if err != nil {
		return nil, err
	}

	_, err = s.io.CreateEgress(ctx, info)
	if err != nil {
		logger.Errorw("failed to create egress", err)
	}

	return info, nil
}

func (s *egressLauncher) StartEgressWithClusterId(ctx context.Context, clusterId string, req *rpc.StartEgressRequest) (*livekit.EgressInfo, error) {
	if s.client == nil {
		return nil, ErrEgressNotConnected
	}

	// Ensure we have an Egress ID
	if req.EgressId == "" {
		req.EgressId = utils.NewGuid(utils.EgressPrefix)
	}

	return s.client.StartEgress(ctx, clusterId, req)
}

type LayoutMetadata struct {
	Layout string `json:"layout"`
}

func (s *EgressService) UpdateLayout(ctx context.Context, req *livekit.UpdateLayoutRequest) (*livekit.EgressInfo, error) {
	AppendLogFields(ctx, "egressID", req.EgressId, "layout", req.Layout)
	if err := EnsureRecordPermission(ctx); err != nil {
		return nil, twirpAuthError(err)
	}

	info, err := s.io.GetEgress(ctx, &rpc.GetEgressRequest{EgressId: req.EgressId})
	if err != nil {
		return nil, err
	}

	metadata, err := json.Marshal(&LayoutMetadata{Layout: req.Layout})
	if err != nil {
		return nil, err
	}

	grants := GetGrants(ctx)
	grants.Video.Room = info.RoomName
	grants.Video.RoomAdmin = true

	_, err = s.roomService.UpdateParticipant(ctx, &livekit.UpdateParticipantRequest{
		Room:     info.RoomName,
		Identity: info.EgressId,
		Metadata: string(metadata),
	})
	if err != nil {
		return nil, err
	}

	return info, nil
}

func (s *EgressService) UpdateStream(ctx context.Context, req *livekit.UpdateStreamRequest) (*livekit.EgressInfo, error) {
	AppendLogFields(ctx, "egressID", req.EgressId, "addUrls", req.AddOutputUrls, "removeUrls", req.RemoveOutputUrls)
	if err := EnsureRecordPermission(ctx); err != nil {
		return nil, twirpAuthError(err)
	}

	if s.client == nil {
		return nil, ErrEgressNotConnected
	}

	info, err := s.client.UpdateStream(ctx, req.EgressId, req)
	if err != nil {
		var loadErr error
		info, loadErr = s.io.GetEgress(ctx, &rpc.GetEgressRequest{EgressId: req.EgressId})
		if loadErr != nil {
			return nil, loadErr
		}

		switch info.Status {
		case livekit.EgressStatus_EGRESS_STARTING,
			livekit.EgressStatus_EGRESS_ACTIVE:
			return nil, err
		default:
			return nil, twirp.NewError(twirp.FailedPrecondition,
				fmt.Sprintf("egress with status %s cannot be updated", info.Status.String()))
		}
	}

	return info, nil
}

func (s *EgressService) UpdateOutputs(ctx context.Context, req *livekit.UpdateOutputsRequest) (*livekit.EgressInfo, error) {
	return nil, twirp.NewError(twirp.Unimplemented, "Update Outputs unimplemented")
}

func (s *EgressService) ListEgress(ctx context.Context, req *livekit.ListEgressRequest) (*livekit.ListEgressResponse, error) {
	if req.RoomName != "" {
		AppendLogFields(ctx, "room", req.RoomName)
	}
	if err := EnsureRecordPermission(ctx); err != nil {
		return nil, twirpAuthError(err)
	}
	return s.io.ListEgress(ctx, req)
}

func (s *EgressService) StopEgress(ctx context.Context, req *livekit.StopEgressRequest) (*livekit.EgressInfo, error) {
	AppendLogFields(ctx, "egressID", req.EgressId)
	if err := EnsureRecordPermission(ctx); err != nil {
		return nil, twirpAuthError(err)
	}

	if s.client == nil {
		return nil, ErrEgressNotConnected
	}

	info, err := s.client.StopEgress(ctx, req.EgressId, req)
	if err != nil {
		var loadErr error
		info, loadErr = s.io.GetEgress(ctx, &rpc.GetEgressRequest{EgressId: req.EgressId})
		if loadErr != nil {
			return nil, loadErr
		}

		switch info.Status {
		case livekit.EgressStatus_EGRESS_STARTING,
			livekit.EgressStatus_EGRESS_ACTIVE:
			return nil, err
		default:
			return nil, twirp.NewError(twirp.FailedPrecondition,
				fmt.Sprintf("egress with status %s cannot be stopped", info.Status.String()))
		}
	}

	return info, nil
}
