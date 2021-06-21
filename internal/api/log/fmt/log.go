package fmt

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/jacklaaa89/pokeapi/internal/api/log"
)

const timeFormat = "2006/01/02 15:04:05"

const (
	// LevelNone sets a logger so show no logging information.
	LevelNone Level = 0

	// LevelError sets a logger to show helpers messages only.
	LevelError Level = 1

	// LevelWarn sets a logger to show warning messages or anything more
	// severe.
	LevelWarn Level = 2

	// LevelInfo sets a logger to show informational messages or anything more
	// severe.
	LevelInfo Level = 3

	// LevelDebug sets a logger to show informational messages or anything more
	// severe.
	LevelDebug Level = 4
)

// Level represents a logging level.
type Level uint32

type logger struct {
	level    Level
	out, err io.Writer
}

// Debugf logs a debug message using Printf conventions.
func (l *logger) Debugf(format string, v ...interface{}) {
	l.log(l.out, LevelDebug, "DEBUG", format, v...)
}

// Errorf logs a warning message using Printf conventions.
func (l *logger) Errorf(format string, v ...interface{}) {
	l.log(l.err, LevelError, "ERROR", format, v...)
}

// Infof logs an informational message using Printf conventions.
func (l *logger) Infof(format string, v ...interface{}) {
	l.log(l.out, LevelInfo, "INFO", format, v...)
}

// Warnf logs a warning message using Printf conventions.
func (l *logger) Warnf(format string, v ...interface{}) {
	l.log(l.err, LevelWarn, "WARN", format, v...)
}

func (l *logger) log(out io.Writer, min Level, prefix, format string, v ...interface{}) {
	if l.level >= min {
		fmt.Fprintf(out, "["+prefix+"] "+time.Now().Format(timeFormat)+" "+format+"\n", v...)
	}
}

// New generates a new logger with the defined minimum logging level
// this will default the outputs to os.Stdout and os.Stderr respectively.
func New(level Level) log.Logger { return NewWithOutputs(level, os.Stdout, os.Stderr) }

// NewWithOutputs generates a new logger with the defined minimum logging level, and
// writers to where logs are written to.
func NewWithOutputs(level Level, out, err io.Writer) log.Logger {
	if level > LevelDebug {
		level = LevelDebug
	}

	return &logger{level, out, err}
}
