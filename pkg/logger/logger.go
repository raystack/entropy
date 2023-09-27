package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LogConfig struct {
	Level string `mapstructure:"level" default:"info"`
}

func Setup(config *LogConfig) error {
	defaultConfig := zap.NewProductionConfig()
	defaultConfig.Level = zap.NewAtomicLevelAt(getZapLogLevelFromString(config.Level))
	logger, err := zap.NewProductionConfig().Build()
	if err != nil {
		return err
	}
	// Setting up global Logger. This can be accessed by zap.L()
	zap.ReplaceGlobals(logger)
	return nil
}

func getZapLogLevelFromString(level string) zapcore.Level {
	l, err := zapcore.ParseLevel(level)
	if err != nil {
		l = zap.InfoLevel
	}
	return l
}
