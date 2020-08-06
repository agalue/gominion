package tools

import (
	"log"
)

// Logger implements basic interface for logging
type Logger struct{}

// Printf implements formatted print
func (logger Logger) Printf(format string, params ...interface{}) {
	log.Printf(format, params...)
}

// Errorf implements formatted error
func (logger Logger) Errorf(format string, params ...interface{}) {
	log.Printf(format, params...)
}

// Warnf implements formatted warning
func (logger Logger) Warnf(format string, params ...interface{}) {
	log.Printf(format, params...)
}

// Debugf implements formatted debug
func (logger Logger) Debugf(format string, params ...interface{}) {
	log.Printf(format, params...)
}

// Infof implements formatted info
func (logger Logger) Infof(format string, params ...interface{}) {
	log.Printf(format, params...)
}

// Fatalf implements formatted fatal
func (logger Logger) Fatalf(format string, params ...interface{}) {
	log.Fatalf(format, params...)
}

// Warn implements warn
func (logger Logger) Warn(...interface{}) {}

// Error implements error
func (logger Logger) Error(...interface{}) {}

// Debug implements debug
func (logger Logger) Debug(...interface{}) {}
