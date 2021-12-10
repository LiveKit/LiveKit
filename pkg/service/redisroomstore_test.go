package service_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	livekit "github.com/livekit/protocol/livekit"
	"github.com/stretchr/testify/require"

	"github.com/livekit/livekit-server/pkg/service"
)

func TestParticipantPersistence(t *testing.T) {
	ctx := context.Background()
	rs := service.NewRedisRoomStore(redisClient())

	roomName := "room1"
	rs.DeleteRoom(ctx, roomName)

	p := &livekit.ParticipantInfo{
		Sid:      "PA_test",
		Identity: "test",
		State:    livekit.ParticipantInfo_ACTIVE,
		Tracks: []*livekit.TrackInfo{
			{
				Sid:  "track1",
				Type: livekit.TrackType_AUDIO,
				Name: "audio",
			},
		},
	}

	// create the participant
	require.NoError(t, rs.StoreParticipant(ctx, roomName, p))

	// result should match
	pGet, err := rs.LoadParticipant(ctx, roomName, p.Identity)
	require.NoError(t, err)
	require.Equal(t, p.Identity, pGet.Identity)
	require.Equal(t, len(p.Tracks), len(pGet.Tracks))
	require.Equal(t, p.Tracks[0].Sid, pGet.Tracks[0].Sid)

	// list should return one participant
	participants, err := rs.ListParticipants(ctx, roomName)
	require.NoError(t, err)
	require.Len(t, participants, 1)

	// deleting participant should return back to normal
	require.NoError(t, rs.DeleteParticipant(ctx, roomName, p.Identity))

	participants, err = rs.ListParticipants(ctx, roomName)
	require.NoError(t, err)
	require.Len(t, participants, 0)

	// shouldn't be able to get it
	_, err = rs.LoadParticipant(ctx, roomName, p.Identity)
	require.Equal(t, err, service.ErrParticipantNotFound)
}

func TestRoomLock(t *testing.T) {
	ctx := context.Background()
	rs := service.NewRedisRoomStore(redisClient())
	lockInterval := 5 * time.Millisecond
	roomName := "myroom"

	t.Run("normal locking", func(t *testing.T) {
		token, err := rs.LockRoom(ctx, roomName, lockInterval)
		require.NoError(t, err)
		require.NotEmpty(t, token)
		require.NoError(t, rs.UnlockRoom(ctx, roomName, token))
	})

	t.Run("waits before acquiring lock", func(t *testing.T) {
		token, err := rs.LockRoom(ctx, roomName, lockInterval)
		require.NoError(t, err)
		require.NotEmpty(t, token)
		unlocked := uint32(0)
		wg := sync.WaitGroup{}

		wg.Add(1)
		go func() {
			// attempt to lock again
			defer wg.Done()
			token2, err := rs.LockRoom(ctx, roomName, lockInterval)
			require.NoError(t, err)
			defer rs.UnlockRoom(ctx, roomName, token2)
			require.Equal(t, uint32(1), atomic.LoadUint32(&unlocked))
		}()

		// release after 2 ms
		time.Sleep(2 * time.Millisecond)
		atomic.StoreUint32(&unlocked, 1)
		rs.UnlockRoom(ctx, roomName, token)

		wg.Wait()
	})

	t.Run("lock expires", func(t *testing.T) {
		token, err := rs.LockRoom(ctx, roomName, lockInterval)
		require.NoError(t, err)
		defer rs.UnlockRoom(ctx, roomName, token)

		time.Sleep(lockInterval + time.Millisecond)
		token2, err := rs.LockRoom(ctx, roomName, lockInterval)
		require.NoError(t, err)
		rs.UnlockRoom(ctx, roomName, token2)
	})
}
