// adapters/logger/zap.go
package logger

import (
	"go.uber.org/zap"

	"github.com/baltikaa9/zxcoin/app/logger"
)

type ZapLogger struct {
	log *zap.SugaredLogger
}

func NewZapLogger() (*ZapLogger, error) {
	l, err := zap.NewDevelopment(zap.AddCallerSkip(1))
	if err != nil {
		return nil, err
	}

	return &ZapLogger{log: l.Sugar()}, nil
}

func (l *ZapLogger) Debug(msg string, kv ...any) { l.log.Debugw(msg, kv...) }
func (l *ZapLogger) Info(msg string, kv ...any)  { l.log.Infow(msg, kv...) }
func (l *ZapLogger) Warn(msg string, kv ...any)  { l.log.Warnw(msg, kv...) }
func (l *ZapLogger) Error(msg string, kv ...any) { l.log.Errorw(msg, kv...) }

func (l *ZapLogger) With(kv ...any) logger.Logger {
	return &ZapLogger{log: l.log.With(kv...)}
}

func (l *ZapLogger) Named(name string) logger.Logger {
	return &ZapLogger{log: l.log.Named(name)}
}

// Sync сбрасывает буфер. Не входит в интерфейс, его вызывает main.
func (l *ZapLogger) Sync() error { return l.log.Sync() }
