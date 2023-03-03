// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package service

import (
	"fmt"
	"github.com/livekit/livekit-server/pkg/clientconfiguration"
	"github.com/livekit/livekit-server/pkg/config"
	"github.com/livekit/livekit-server/pkg/routing"
	"github.com/livekit/livekit-server/pkg/telemetry"
	"github.com/livekit/protocol/auth"
	"github.com/livekit/protocol/egress"
	"github.com/livekit/protocol/livekit"
	redis2 "github.com/livekit/protocol/redis"
	"github.com/livekit/protocol/rpc"
	"github.com/livekit/protocol/utils"
	"github.com/livekit/protocol/webhook"
	"github.com/livekit/psrpc"
	"github.com/pion/turn/v2"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"gopkg.in/yaml.v3"
	"os"
)

import (
	_ "net/http/pprof"
)

// Injectors from wire.go:

func InitializeServer(conf *config.Config, currentNode routing.LocalNode) (*LivekitServer, error) {
	roomConfig := getRoomConf(conf)
	apiConfig := config.DefaultAPIConfig()
	universalClient, err := createRedisClient(conf)
	if err != nil {
		return nil, err
	}
	nodeID := getNodeID(currentNode)
	messageBus := getMessageBus(universalClient)
	signalClient, err := routing.NewSignalClient(nodeID, messageBus)
	if err != nil {
		return nil, err
	}
	clientConfig := getClientConfig(conf)
	router := routing.CreateRouter(universalClient, currentNode, signalClient, clientConfig)
	objectStore := createStore(universalClient)
	roomAllocator, err := NewRoomAllocator(conf, router, objectStore)
	if err != nil {
		return nil, err
	}
	egressClient, err := getEgressClient(conf, nodeID, messageBus)
	if err != nil {
		return nil, err
	}
	rpcClient := egress.NewRedisRPCClient(nodeID, universalClient)
	egressStore := getEgressStore(objectStore)
	keyProvider, err := createKeyProvider(conf)
	if err != nil {
		return nil, err
	}
	notifier, err := createWebhookNotifier(conf, keyProvider)
	if err != nil {
		return nil, err
	}
	analyticsService := telemetry.NewAnalyticsService(conf, currentNode)
	telemetryService := telemetry.NewTelemetryService(notifier, analyticsService)
	rtcEgressLauncher := NewEgressLauncher(egressClient, rpcClient, egressStore, telemetryService)
	roomService, err := NewRoomService(roomConfig, apiConfig, router, roomAllocator, objectStore, rtcEgressLauncher)
	if err != nil {
		return nil, err
	}
	egressService := NewEgressService(egressClient, rpcClient, objectStore, egressStore, roomService, telemetryService, rtcEgressLauncher)
	ingressConfig := getIngressConfig(conf)
	ingressClient, err := rpc.NewIngressClient(nodeID, messageBus)
	if err != nil {
		return nil, err
	}
	ingressStore := getIngressStore(objectStore)
	ingressService := NewIngressService(ingressConfig, nodeID, messageBus, ingressClient, ingressStore, roomService, telemetryService)
	ioInfoService, err := NewIOInfoService(nodeID, messageBus, egressStore, ingressStore, telemetryService, rpcClient)
	if err != nil {
		return nil, err
	}
	rtcService := NewRTCService(conf, roomAllocator, objectStore, router, currentNode, telemetryService)
	clientConfigurationManager := createClientConfiguration()
	timedVersionGenerator := utils.NewDefaultTimedVersionGenerator()
	roomManager, err := NewLocalRoomManager(conf, objectStore, currentNode, router, telemetryService, clientConfigurationManager, rtcEgressLauncher, timedVersionGenerator)
	if err != nil {
		return nil, err
	}
	signalServer, err := NewDefaultSignalServer(currentNode, messageBus, router, roomManager)
	if err != nil {
		return nil, err
	}
	authHandler := newTurnAuthHandler(objectStore)
	server, err := newInProcessTurnServer(conf, authHandler)
	if err != nil {
		return nil, err
	}
	livekitServer, err := NewLivekitServer(conf, roomService, egressService, ingressService, ioInfoService, rtcService, keyProvider, router, roomManager, signalServer, server, currentNode)
	if err != nil {
		return nil, err
	}
	return livekitServer, nil
}

func InitializeRouter(conf *config.Config, currentNode routing.LocalNode) (routing.Router, error) {
	universalClient, err := createRedisClient(conf)
	if err != nil {
		return nil, err
	}
	nodeID := getNodeID(currentNode)
	messageBus := getMessageBus(universalClient)
	signalClient, err := routing.NewSignalClient(nodeID, messageBus)
	if err != nil {
		return nil, err
	}
	clientConfig := getClientConfig(conf)
	router := routing.CreateRouter(universalClient, currentNode, signalClient, clientConfig)
	return router, nil
}

// wire.go:

func getNodeID(currentNode routing.LocalNode) livekit.NodeID {
	return livekit.NodeID(currentNode.Id)
}

func createKeyProvider(conf *config.Config) (auth.KeyProvider, error) {

	if conf.KeyFile != "" {
		if st, err := os.Stat(conf.KeyFile); err != nil {
			return nil, err
		} else if st.Mode().Perm() != 0600 {
			return nil, fmt.Errorf("key file must have permission set to 600")
		}
		f, err := os.Open(conf.KeyFile)
		if err != nil {
			return nil, err
		}
		defer func() {
			_ = f.Close()
		}()
		decoder := yaml.NewDecoder(f)
		if err = decoder.Decode(conf.Keys); err != nil {
			return nil, err
		}
	}

	if len(conf.Keys) == 0 {
		return nil, errors.New("one of key-file or keys must be provided in order to support a secure installation")
	}

	return auth.NewFileBasedKeyProviderFromMap(conf.Keys), nil
}

func createWebhookNotifier(conf *config.Config, provider auth.KeyProvider) (webhook.Notifier, error) {
	wc := conf.WebHook
	if len(wc.URLs) == 0 {
		return nil, nil
	}
	secret := provider.GetSecret(wc.APIKey)
	if secret == "" {
		return nil, ErrWebHookMissingAPIKey
	}

	return webhook.NewNotifier(wc.APIKey, secret, wc.URLs), nil
}

func createRedisClient(conf *config.Config) (redis.UniversalClient, error) {
	if !conf.Redis.IsConfigured() {
		return nil, nil
	}
	return redis2.GetRedisClient(&conf.Redis)
}

func createStore(rc redis.UniversalClient) ObjectStore {
	if rc != nil {
		return NewRedisStore(rc)
	}
	return NewLocalStore()
}

func getMessageBus(rc redis.UniversalClient) psrpc.MessageBus {
	if rc == nil {
		return psrpc.NewLocalMessageBus()
	}
	return psrpc.NewRedisMessageBus(rc)
}

func getEgressClient(conf *config.Config, nodeID livekit.NodeID, bus psrpc.MessageBus) (rpc.EgressClient, error) {
	if conf.Egress.UsePsRPC {
		return rpc.NewEgressClient(nodeID, bus)
	}

	return nil, nil
}

func getEgressStore(s ObjectStore) EgressStore {
	switch store := s.(type) {
	case *RedisStore:
		return store
	default:
		return nil
	}
}

func getIngressStore(s ObjectStore) IngressStore {
	switch store := s.(type) {
	case *RedisStore:
		return store
	default:
		return nil
	}
}

func getIngressConfig(conf *config.Config) *config.IngressConfig {
	return &conf.Ingress
}

func createClientConfiguration() clientconfiguration.ClientConfigurationManager {
	return clientconfiguration.NewStaticClientConfigurationManager(clientconfiguration.StaticConfigurations)
}

func getRoomConf(config2 *config.Config) config.RoomConfig {
	return config2.Room
}

func newInProcessTurnServer(conf *config.Config, authHandler turn.AuthHandler) (*turn.Server, error) {
	return NewTurnServer(conf, authHandler, false)
}

func getClientConfig(config2 *config.Config) config.ClientConfig {
	return config2.Clients
}
