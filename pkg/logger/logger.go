package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// GetProductionLogger initialises production environment logger
func GetProductionLogger(appName string, appVersion string) (*zap.SugaredLogger, error) {
	config := zap.NewProductionConfig()
	config.OutputPaths = []string{"stdout"}
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.DisableStacktrace = true
	config.InitialFields = map[string]interface{}{
		"service": appName,
		"version": appVersion,
	}
	lb, err := config.Build()
	if err != nil {
		return nil, err
	}
	return lb.Sugar(), nil
}
