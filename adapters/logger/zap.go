package logger

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/baltikaa9/zxcoin/app/logger"
)

var _ logger.Logger = (*ZapLogger)(nil)

type ZapLogger struct {
	log *zap.SugaredLogger
}

func NewZapLogger(path string) (*ZapLogger, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)

	if err != nil {
		return nil, err
	}

	consoleCfg := zap.NewDevelopmentEncoderConfig()
	consoleCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	consoleCfg.EncodeTime = zapcore.TimeEncoderOfLayout("15:04:05.000")

	fileCfg := zap.NewProductionEncoderConfig()
	fileCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	core := zapcore.NewTee(
		zapcore.NewCore(zapcore.NewConsoleEncoder(consoleCfg), zapcore.Lock(os.Stdout), zapcore.DebugLevel),
		zapcore.NewCore(zapcore.NewJSONEncoder(fileCfg), zapcore.Lock(file), zapcore.InfoLevel),
	)

	log := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1), zap.AddStacktrace(zapcore.ErrorLevel))

	return &ZapLogger{log: log.Sugar()}, nil
}

func (l *ZapLogger) Debug(msg string, kv ...any) {
	l.log.Debugw(msg, kv...)
}

func (l *ZapLogger) Info(msg string, kv ...any) {
	l.log.Infow(msg, kv...)
}

func (l *ZapLogger) Warn(msg string, kv ...any) {
	l.log.Warnw(msg, kv...)
}

func (l *ZapLogger) Error(msg string, kv ...any) {
	l.log.Errorw(msg, kv...)
}

func (l *ZapLogger) With(kv ...any) logger.Logger {
	return &ZapLogger{log: l.log.With(kv...)}
}

func (l *ZapLogger) Named(name string) logger.Logger {
	return &ZapLogger{log: l.log.Named(name)}
}

func (l *ZapLogger) Sync() error {
	return l.log.Sync()
}
