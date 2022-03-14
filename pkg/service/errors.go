package service

import "errors"

var (
	ErrEgressNotFound       = errors.New("egress does not exist")
	ErrEgressNotConnected   = errors.New("egress not connected (redis required)")
	ErrRoomNotFound         = errors.New("requested room does not exist")
	ErrRoomLockFailed       = errors.New("could not lock room")
	ErrRoomUnlockFailed     = errors.New("could not unlock room, lock token does not match")
	ErrParticipantNotFound  = errors.New("participant does not exist")
	ErrTrackNotFound        = errors.New("track is not found")
	ErrWebHookMissingAPIKey = errors.New("api_key is required to use webhooks")
	ErrOperationFailed      = errors.New("operation cannot be completed")
)
