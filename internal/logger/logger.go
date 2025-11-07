package logger

import (
	"context"
	"fmt"
	"go.opentelemetry.io/contrib/bridges/otellogrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"time"
)

type Logger struct {
	zap *zap.Logger
}

func New(env string, serviceName string) (*Logger, error) {
	var cfg zap.Config
	switch env {
	case "development":
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.TimeKey = "timestamp"
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		cfg.EncoderConfig.CallerKey = "caller"
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	case "staging", "production":
		cfg = zap.NewProductionConfig()
		cfg.EncoderConfig.TimeKey = "timestamp"
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	default:
		return nil, fmt.Errorf("unknown environment: %s", env)
	}

	cfg.OutputPaths = []string{"stdout", fmt.Sprintf("./logs/%s.log", serviceName)}

	zapLogger, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	// Link to OTEL logs if available
	provider := otel.GetTracerProvider()
	if provider != nil {
		zapLogger.Info("OpenTelemetry tracer provider detected")
	}

	return &Logger{zap: zapLogger}, nil
}

func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.zap.Info(msg, fields...)
}

func (l *Logger) Error(msg string, err error, fields ...zap.Field) {
	l.zap.Error(msg, append(fields, zap.Error(err))...)
}

func (l *Logger) Sync() {
	_ = l.zap.Sync()
}

func (l *Logger) Trace(ctx context.Context, msg string, fields ...zap.Field) {
	span := otel.Tracer("logger").Start(ctx, msg)
	defer span.End()
	span.SpanContext()
	l.zap.Info(msg, fields...)
	span.SetAttributes(attribute.String("log.message", msg))
	span.SetAttributes(attribute.String("timestamp", time.Now().Format(time.RFC3339)))
}
