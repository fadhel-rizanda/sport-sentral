package logger

import (
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	instance *zap.Logger
	once     sync.Once
)

// New returns a singleton zap.Logger.
// Call once at startup: logger.New("production")
func New(env string) *zap.Logger {
	once.Do(func() {
		instance = build(env)
	})
	return instance
}

// Get returns the singleton. Panics if New was not called first.
func Get() *zap.Logger {
	if instance == nil {
		panic("logger not initialized: call logger.New() first")
	}
	return instance
}

func build(env string) *zap.Logger {
	var cfg zap.Config

	if env == "production" {
		cfg = zap.NewProductionConfig()
		cfg.EncoderConfig.TimeKey = "ts"
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// Allow LOG_LEVEL env override
	if lvl := os.Getenv("LOG_LEVEL"); lvl != "" {
		var l zapcore.Level
		if err := l.UnmarshalText([]byte(lvl)); err == nil {
			cfg.Level = zap.NewAtomicLevelAt(l)
		}
	}

	log, err := cfg.Build(zap.AddCallerSkip(0))
	if err != nil {
		panic("failed to build logger: " + err.Error())
	}
	return log
}
