package log

// Logger provides a basic leveled logging interface for
// printing debug, informational, warning, and helpers messages.
type Logger interface {
	// Debugf logs a debug message using Printf conventions.
	Debugf(format string, v ...interface{})
	// Errorf logs a warning message using Printf conventions.
	Errorf(format string, v ...interface{})
	// Infof logs an informational message using Printf conventions.
	Infof(format string, v ...interface{})
	// Warnf logs a warning message using Printf conventions.
	Warnf(format string, v ...interface{})
}
