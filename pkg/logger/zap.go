package logger

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ZapLogger struct {
	*zap.Logger
}

func (z *ZapLogger) Debug(msg string, fields ...zap.Field) {
	z.Logger.Debug(msg, fields...)
}

func (z *ZapLogger) Info(msg string, fields ...zap.Field) {
	z.Logger.Info(msg, fields...)
}

func (z *ZapLogger) Warn(msg string, fields ...zap.Field) {
	z.Logger.Warn(msg, fields...)
}

func (z *ZapLogger) Error(msg string, fields ...zap.Field) {
	z.Logger.Error(msg, fields...)
}

func (z *ZapLogger) Panic(msg string, fields ...zap.Field) {
	z.Logger.Panic(msg, fields...)
}

func (z *ZapLogger) Fatal(msg string, fields ...zap.Field) {
	z.Logger.Fatal(msg, fields...)
}

func NewLogger(logLevel string) (*ZapLogger, error) {
	if logLevel == "none" {
		return &ZapLogger{zap.NewNop()}, nil
	}

	var level zapcore.Level
	switch logLevel {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	case "panic":
		level = zapcore.PanicLevel
	case "fatal":
		level = zapcore.FatalLevel
	default:
		return nil, fmt.Errorf("unknown logger level: %s", logLevel)
	}

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.MessageKey = "message"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	config := zap.Config{
		Level:             zap.NewAtomicLevelAt(level),
		Development:       false,
		DisableCaller:     false,
		DisableStacktrace: true,
		Sampling:          nil,
		Encoding:          "json",
		EncoderConfig:     encoderCfg,
		OutputPaths:       []string{"stderr"},
		ErrorOutputPaths:  []string{"stderr"},
		InitialFields:     map[string]any{"pid": os.Getpid()},
	}

	logger, err := config.Build()
	if err != nil {
		return nil, err
	}

	logger = logger.WithOptions(zap.AddCallerSkip(1))
	return &ZapLogger{logger}, nil
}

func MustNewLogger(logLevel string) *ZapLogger {
	l, err := NewLogger(logLevel)
	if err != nil {
		panic(err)
	}

	return l
}
