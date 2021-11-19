package log

import (
	"fmt"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger
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

// GetLogger returns the zap logger
func GetLogger() *zap.Logger {
	return logger
}

// GetSugaredLogger returns the sugared zap logger
func GetSugaredLogger() *zap.SugaredLogger {
	return log
}

// InitLogger initializes colorized logger
func InitLogger(logLevel string) {
	level := getLogLevel(logLevel)
	config := zap.Config{
		Level:             level,
		Development:       false,
		DisableStacktrace: true,
		DisableCaller:     true,
		Encoding:          "console",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "logger",
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
	var err error
	logger, err = config.Build()
	if err != nil {
		panic(err)
	}
	log = logger.Sugar()
}

// InitProdLogger initializes production logger
func InitProdLogger(logLevel string) {
	config := zap.NewProductionConfig()
	config.Level = getLogLevel(logLevel)
	config.DisableStacktrace = true
	config.DisableCaller = true
	var err error
	logger, err = config.Build()
	if err != nil {
		panic(err)
	}
	log = logger.Sugar()
}

func getLogLevel(logLevel string) zap.AtomicLevel {
	level := zap.NewAtomicLevel()
	switch strings.ToLower(logLevel) {
	case "debug":
		level.SetLevel(zap.DebugLevel)
	case "info":
		level.SetLevel(zap.InfoLevel)
	case "warn":
		level.SetLevel(zap.WarnLevel)
	case "error":
		level.SetLevel(zap.ErrorLevel)
	}
	fmt.Printf("Logging level: %s\n", level.String())
	return level
}
