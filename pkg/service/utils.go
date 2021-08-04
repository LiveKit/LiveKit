package service

import (
	"context"
	"net/http"

	"github.com/go-redis/redis/v8"
	"github.com/google/wire"
	"github.com/pkg/errors"

	"github.com/livekit/livekit-server/pkg/config"
	"github.com/livekit/livekit-server/pkg/logger"
	"github.com/livekit/livekit-server/pkg/routing"
	livekit "github.com/livekit/livekit-server/proto"
)

var ServiceSet = wire.NewSet(
	createRedisClient,
	createRouter,
	createStore,
	NewRecordingService,
	NewRoomService,
	NewRTCService,
	NewLivekitServer,
	NewRoomManager,
	NewTurnServer,
	config.GetAudioConfig,
	wire.Bind(new(livekit.RecordingService), new(*RecordingService)),
	wire.Bind(new(livekit.RoomService), new(*RoomService)),
)

func createRedisClient(conf *config.Config) (*redis.Client, error) {
	if !conf.HasRedis() {
		return nil, nil
	}

	logger.Infow("using multi-node routing via redis", "addr", conf.Redis.Address)
	rc := redis.NewClient(&redis.Options{
		Addr:     conf.Redis.Address,
		Username: conf.Redis.Username,
		Password: conf.Redis.Password,
		DB:       conf.Redis.DB,
	})
	if err := rc.Ping(context.Background()).Err(); err != nil {
		err = errors.Wrap(err, "unable to connect to redis")
		return nil, err
	}

	return rc, nil
}

func createRouter(rc *redis.Client, node routing.LocalNode) routing.Router {
	if rc != nil {
		return routing.NewRedisRouter(node, rc)
	}

	// local routing and store
	logger.Infow("using single-node routing")
	return routing.NewLocalRouter(node)
}

func createStore(rc *redis.Client) RoomStore {
	if rc != nil {
		return NewRedisRoomStore(rc)
	}
	return NewLocalRoomStore()
}

func handleError(w http.ResponseWriter, status int, msg string) {
	// GetLogger already with extra depth 1
	logger.GetLogger().V(1).Info("error handling request", "error", msg, "status", status)
	w.WriteHeader(status)
	_, _ = w.Write([]byte(msg))
}

func boolValue(s string) bool {
	return s == "1" || s == "true"
}
