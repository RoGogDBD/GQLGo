package config

import (
	"go.uber.org/zap"

	"github.com/RoGogDBD/GQLGo/internal/service"
)

type printfLogger struct {
	*zap.SugaredLogger
}

func (l printfLogger) Infof(format string, args ...any)  { l.SugaredLogger.Infof(format, args...) }
func (l printfLogger) Warnf(format string, args ...any)  { l.SugaredLogger.Warnf(format, args...) }
func (l printfLogger) Errorf(format string, args ...any) { l.SugaredLogger.Errorf(format, args...) }

// NewLogger инициализация.
func NewLogger() (service.Logger, func(), error) {
	zapLogger, err := zap.NewDevelopment()
	if err != nil {
		return nil, func() {}, err
	}
	cleanup := func() {
		_ = zapLogger.Sync()
	}
	return printfLogger{zapLogger.Sugar()}, cleanup, nil
}
