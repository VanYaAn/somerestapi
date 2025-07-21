package logger

import (
	"fmt"

	"go.uber.org/zap"
)

func NewLogger(level string) (*zap.SugaredLogger, error) {
	var zapConfig zap.Config
	switch level {
	case "debug":
		zapConfig = zap.NewDevelopmentConfig()
	case "info":
		zapConfig = zap.NewProductionConfig()
	default:
		return nil, fmt.Errorf("unsupported log level: %s", level)
	}

	logger, err := zapConfig.Build()
	if err != nil {
		return nil, fmt.Errorf("error building logger: %w", err)
	}

	return logger.Sugar(), nil
}
