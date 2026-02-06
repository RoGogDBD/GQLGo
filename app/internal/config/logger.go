package config

import (
	"go.uber.org/zap"

	"github.com/RoGogDBD/GQLGo/internal/logging"
)

type Logger struct {
	*zap.SugaredLogger
}

func (l Logger) Infof(format string, args ...any)  { l.SugaredLogger.Infof(format, args...) }
func (l Logger) Warnf(format string, args ...any)  { l.SugaredLogger.Warnf(format, args...) }
func (l Logger) Errorf(format string, args ...any) { l.SugaredLogger.Errorf(format, args...) }

func NewLogger() (logging.Logger, func(), error) {
	zapLogger, err := zap.NewDevelopment()
	if err != nil {
		return nil, func() {}, err
	}
	cleanup := func() {
		_ = zapLogger.Sync()
	}
	return Logger{zapLogger.Sugar()}, cleanup, nil
}
