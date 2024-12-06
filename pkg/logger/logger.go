package logger

import (
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

type Logger struct {
    *zap.Logger
}

func NewLogger() *Logger {
    config := zap.NewProductionConfig()
    config.EncoderConfig.TimeKey = "timestamp"
    config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
    
    logger, err := config.Build()
    if err != nil {
        panic(err)
    }
    
    return &Logger{Logger: logger}
}

func (l *Logger) Fatal(msg string, err error) {
    l.Logger.Fatal(msg, zap.Error(err))
} 