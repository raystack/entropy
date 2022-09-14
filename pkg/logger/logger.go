package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LogConfig struct {
	Level string `mapstructure:"level" default:"info"`
}

func New(config *LogConfig) (*zap.Logger, error) {
	defaultConfig := zap.NewProductionConfig()
	defaultConfig.Level = zap.NewAtomicLevelAt(getZapLogLevelFromString(config.Level))
	logger, err := zap.NewProductionConfig().Build()
	return logger, err
}

func getZapLogLevelFromString(level string) zapcore.Level {
	l, err := zapcore.ParseLevel(level)
	if err != nil {
		l = zap.InfoLevel
	}
	return l
}
