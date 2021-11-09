package serverlogger

import (
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/livekit/livekit-server/pkg/sfu"
	"github.com/livekit/livekit-server/pkg/sfu/buffer"
	"github.com/livekit/protocol/logger"
	"github.com/pion/logging"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// pion/webrtc, pion/turn
	defaultFactory logging.LoggerFactory
)

func LoggerFactory() logging.LoggerFactory {
	if defaultFactory == nil {
		defaultFactory = logging.NewDefaultLoggerFactory()
	}
	return defaultFactory
}

func SetLoggerFactory(lf logging.LoggerFactory) {
	defaultFactory = lf
}

// Note: only pass in logr.Logger with default depth
func SetLogger(l logr.Logger) {
	logger.SetLogger(l, "livekit")
	sfu.Logger = l.WithName("sfu")
	buffer.Logger = sfu.Logger
}

func InitProduction(logLevel string) {
	initLogger(zap.NewProductionConfig(), logLevel)
}

func InitDevelopment(logLevel string) {
	initLogger(zap.NewDevelopmentConfig(), logLevel)
}

// valid levels: debug, info, warn, error, fatal, panic
func initLogger(config zap.Config, level string) {
	if level != "" {
		lvl := zapcore.Level(0)
		if err := lvl.UnmarshalText([]byte(level)); err == nil {
			config.Level = zap.NewAtomicLevelAt(lvl)
		}
	}

	l, _ := config.Build()
	zapLogger := zapr.NewLogger(l)
	SetLogger(zapLogger)
}
