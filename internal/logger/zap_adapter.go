package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ZapAdapter struct {
	logger *zap.Logger
}

func NewZapAdapter() (*ZapAdapter, error) {
	var config zap.Config
	if os.Getenv("APP_ENV") == "production" {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}

	zapLogger, err := config.Build()
	if err != nil {
		return nil, err
	}
	return &ZapAdapter{logger: zapLogger}, nil
}

func (z *ZapAdapter) Debug(msg string, fields ...Field) {
	z.logger.Debug(msg, convertFields(fields)...)
}

func (z *ZapAdapter) Info(msg string, fields ...Field) {
	z.logger.Info(msg, convertFields(fields)...)
}

func (z *ZapAdapter) Warn(msg string, fields ...Field) {
	z.logger.Warn(msg, convertFields(fields)...)
}

func (z *ZapAdapter) Error(msg string, fields ...Field) {
	z.logger.Error(msg, convertFields(fields)...)
}

func (z *ZapAdapter) Sync() error {
	return z.logger.Sync()
}

func convertFields(fields []Field) []zap.Field {
	zapFields := make([]zap.Field, 0, len(fields))
	for _, f := range fields {
		zapFields = append(zapFields, zap.Any(f.Key, f.Value))
	}
	return zapFields
}
