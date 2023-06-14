// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package service

import (
	"context"
	"github.com/dTelecom/p2p-realtime-database"
	"github.com/ipfs/go-log/v2"
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
	signalRelayConfig := getSignalRelayConfig(conf)
	signalClient, err := routing.NewSignalClient(nodeID, messageBus, signalRelayConfig)
	if err != nil {
		return nil, err
	}
	router := routing.CreateRouter(conf, universalClient, currentNode, signalClient)
	p2p_databaseConfig := getDatabaseConfiguration(conf)
	db, err := createMainDatabaseP2P(p2p_databaseConfig)
	if err != nil {
		return nil, err
	}
	participantCounter := createParticipantCounter(db, nodeID)
	objectStore := createStore(p2p_databaseConfig, nodeID, participantCounter)
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
	keyProviderPublicKey, err := createKeyPublicKeyProvider(conf)
	if err != nil {
		return nil, err
	}
	clientConfigurationManager := createClientConfiguration()
	timedVersionGenerator := utils.NewDefaultTimedVersionGenerator()
	roomManager, err := NewLocalRoomManager(conf, objectStore, currentNode, router, telemetryService, clientConfigurationManager, rtcEgressLauncher, timedVersionGenerator)
	if err != nil {
		return nil, err
	}
	signalServer, err := NewDefaultSignalServer(currentNode, messageBus, signalRelayConfig, router, roomManager)
	if err != nil {
		return nil, err
	}
	authHandler := newTurnAuthHandler(objectStore)
	server, err := newInProcessTurnServer(conf, authHandler)
	if err != nil {
		return nil, err
	}
	livekitServer, err := NewLivekitServer(conf, roomService, egressService, ingressService, ioInfoService, rtcService, keyProviderPublicKey, router, roomManager, signalServer, server, currentNode)
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
	signalRelayConfig := getSignalRelayConfig(conf)
	signalClient, err := routing.NewSignalClient(nodeID, messageBus, signalRelayConfig)
	if err != nil {
		return nil, err
	}
	router := routing.CreateRouter(conf, universalClient, currentNode, signalClient)
	return router, nil
}

// wire.go:

func createParticipantCounter(mainDatabase *p2p_database.DB, nodeId livekit.NodeID) *ParticipantCounter {
	return NewParticipantCounter(nodeId, mainDatabase)
}

func getDatabaseConfiguration(conf *config.Config) p2p_database.Config {
	return p2p_database.Config{
		PeerListenPort:          conf.Ethereum.P2pNodePort,
		EthereumNetworkHost:     conf.Ethereum.NetworkHost,
		EthereumNetworkKey:      conf.Ethereum.NetworkKey,
		EthereumContractAddress: conf.Ethereum.ContractAddress,
		WalletPrivateKey:        conf.Ethereum.WalletPrivateKey,
		DatabaseName:            conf.Ethereum.P2pMainDatabaseName,
	}
}

func createMainDatabaseP2P(conf p2p_database.Config) (*p2p_database.DB, error) {
	db, err := p2p_database.Connect(context.Background(), conf, log.Logger("db"))
	if err != nil {
		return nil, errors.Wrap(err, "create main p2p db")
	}
	return db, nil
}

func getNodeID(currentNode routing.LocalNode) livekit.NodeID {
	return livekit.NodeID(currentNode.Id)
}

func createKeyProvider(conf *config.Config) (auth.KeyProvider, error) {
	return createKeyPublicKeyProvider(conf)
}

func createKeyPublicKeyProvider(conf *config.Config) (auth.KeyProviderPublicKey, error) {
	contract, err := p2p_database.NewEthSmartContract(p2p_database.Config{
		EthereumNetworkHost:     conf.Ethereum.NetworkHost,
		EthereumNetworkKey:      conf.Ethereum.NetworkKey,
		EthereumContractAddress: conf.Ethereum.ContractAddress,
	}, nil)

	if err != nil {
		return nil, errors.Wrap(err, "try create contract")
	}

	return auth.NewEthKeyProvider(*contract, conf.Ethereum.WalletAddress, conf.Ethereum.WalletPrivateKey), nil
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

func createStore(p2pDbConfig p2p_database.Config, nodeID livekit.NodeID, participantCounter *ParticipantCounter) ObjectStore {
	return NewLocalStore(nodeID, p2pDbConfig, participantCounter)
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

func getSignalRelayConfig(config2 *config.Config) config.SignalRelayConfig {
	return config2.SignalRelay
}

func newInProcessTurnServer(conf *config.Config, authHandler turn.AuthHandler) (*turn.Server, error) {
	return NewTurnServer(conf, authHandler, false)
}
