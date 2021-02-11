package auth_test

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/square/go-jose.v2/jwt"

	"github.com/livekit/livekit-server/pkg/auth"
	"github.com/livekit/livekit-server/pkg/utils"
)

func TestAccessToken(t *testing.T) {
	t.Parallel()

	t.Run("keys must be set", func(t *testing.T) {
		token := auth.NewAccessToken("", "")
		_, err := token.ToJWT()
		assert.Equal(t, auth.ErrKeysMissing, err)
	})

	t.Run("generates a decodeable key", func(t *testing.T) {
		apiKey, secret := apiKeypair()
		videoGrant := &auth.VideoGrant{RoomJoin: true, Room: "myroom"}
		at := auth.NewAccessToken(apiKey, secret).
			AddGrant(videoGrant).
			SetValidFor(time.Minute * 5).
			SetIdentity("user")
		value, err := at.ToJWT()
		//fmt.Println(raw)
		assert.NoError(t, err)

		assert.Len(t, strings.Split(value, "."), 3)

		// ensure it's a valid JWT
		token, err := jwt.ParseSigned(value)
		assert.NoError(t, err)

		decodedGrant := auth.ClaimGrants{}
		err = token.UnsafeClaimsWithoutVerification(&decodedGrant)
		assert.NoError(t, err)

		assert.EqualValues(t, videoGrant, decodedGrant.Video)
	})

	t.Run("default validity should be more than a minute", func(t *testing.T) {
		apiKey, secret := apiKeypair()
		videoGrant := &auth.VideoGrant{RoomJoin: true, Room: "myroom"}
		at := auth.NewAccessToken(apiKey, secret).
			AddGrant(videoGrant)
		value, err := at.ToJWT()
		token, err := jwt.ParseSigned(value)

		claim := jwt.Claims{}
		decodedGrant := auth.ClaimGrants{}
		err = token.UnsafeClaimsWithoutVerification(&claim, &decodedGrant)
		assert.NoError(t, err)
		assert.EqualValues(t, videoGrant, decodedGrant.Video)

		// default validity
		assert.True(t, claim.Expiry.Time().Sub(claim.IssuedAt.Time()) > time.Minute)
	})
}

func apiKeypair() (string, string) {
	return utils.NewGuid(utils.APIKeyPrefix), utils.RandomSecret()
}
