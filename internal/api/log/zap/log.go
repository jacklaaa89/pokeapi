package zap

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/jacklaaa89/pokeapi/internal/api/log"
)

type logger struct{ l *zap.Logger }

// Debugf logs a debug message using Printf conventions.
func (l *logger) Debugf(format string, v ...interface{}) {
	l.l.Debug(fmt.Sprintf(format, v...))
}

// Errorf logs a warning message using Printf conventions.
func (l *logger) Errorf(format string, v ...interface{}) {
	l.l.Error(fmt.Sprintf(format, v...))
}

// Infof logs an informational message using Printf conventions.
func (l *logger) Infof(format string, v ...interface{}) {
	l.l.Info(fmt.Sprintf(format, v...))
}

// Warnf logs a warning message using Printf conventions.
func (l *logger) Warnf(format string, v ...interface{}) {
	l.l.Warn(fmt.Sprintf(format, v...))
}

// New generates a new logger with using the supplied zap config.
func New(c *zap.Config, o ...zap.Option) (log.Logger, error) {
	l, err := c.Build(o...)
	return &logger{l}, err
}
