package log

import (
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.SugaredLogger

// Fatalf logs a formatted fatal message
func Fatalf(format string, params ...interface{}) {
	if log != nil {
		log.Fatalf(format, params...)
	}
}

// Errorf logs a formatted error message
func Errorf(format string, params ...interface{}) {
	if log != nil {
		log.Errorf(format, params...)
	}
}

// Warnf logs a formatted warn message
func Warnf(format string, params ...interface{}) {
	if log != nil {
		log.Warnf(format, params...)
	}
}

// Infof logs a formatted info message
func Infof(format string, params ...interface{}) {
	if log != nil {
		log.Infof(format, params...)
	}
}

// Debugf logs a formatted debug message
func Debugf(format string, params ...interface{}) {
	if log != nil {
		log.Debugf(format, params...)
	}
}

// InitLogger initializes the error logger
func InitLogger(logLevel string) {
	level := zap.NewAtomicLevel()
	switch strings.ToLower(logLevel) {
	case "debug":
		level.SetLevel(zap.DebugLevel)
	case "info":
		level.SetLevel(zap.InfoLevel)
	case "warn":
		level.SetLevel(zap.InfoLevel)
	case "error":
		level.SetLevel(zap.ErrorLevel)
	}
	config := zap.Config{
		Level:             level,
		Development:       false,
		DisableStacktrace: true,
		Encoding:          "console",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.CapitalColorLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}
	logger, err := config.Build()
	if err != nil {
		panic(err)
	}
	log = logger.Sugar()
}

func init() {
	if log == nil {
		InitLogger("debug")
	}
}
