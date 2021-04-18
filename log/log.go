// This package is a wrapper for log functions
// in order to ensure substitutability with other libraries
package log

import "github.com/sirupsen/logrus"

// Level type
type Level uint8

// These are the different logging levels.
const (
	// PanicLevel level, highest level of severity. Logs and then calls panic with the
	// message passed to Debug, Info, ...
	PanicLevel Level = iota
	// FatalLevel level. Logs and then calls `os.Exit(1)`. It will exit even if the
	// logging level is set to Panic.
	FatalLevel
	// ErrorLevel level. Logs. Used for errors that should definitely be noted.
	// Commonly used for hooks to send errors to an error tracking service.
	ErrorLevel
	// WarnLevel level. Non-critical entries that deserve eyes.
	WarnLevel
	// InfoLevel level. General operational entries about what's going on inside the
	// application.
	InfoLevel
	// DebugLevel level. Usually only enabled when debugging. Very verbose logging.
	DebugLevel
)

// Timestamp prints logs with timestam
func Timestamp() {
	customFormatter := new(logrus.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	customFormatter.FullTimestamp = true
	logrus.SetFormatter(customFormatter)
}

// SetLevel sets the standard logger level.
func SetLevel(level Level) {
	logrus.SetLevel(logrus.Level(level))
}

func Printf(format string, v ...interface{}) {
	logrus.Printf(format, v...)
}

func Println(v ...interface{}) {
	logrus.Println(v...)
}

func Warnln(args ...interface{}) {
	logrus.Warnln(args...)
}

func Warnf(format string, args ...interface{}) {
	logrus.Warnf(format, args...)
}

func Errorln(args ...interface{}) {
	logrus.Errorln(args...)
}

func Errorf(format string, args ...interface{}) {
	logrus.Errorf(format, args...)
}

func Fatalf(format string, v ...interface{}) {
	logrus.Fatalf(format, v...)
}

func Fatal(v ...interface{}) {
	logrus.Fatal(v...)
}

func Debugf(format string, v ...interface{}) {
	logrus.Debugf(format, v...)
}

type Fields map[string]interface{}

// WithFields creates an entry from the standard logger and adds multiple
// fields to it. This is simply a helper for `WithField`, invoking it
// once for each field.
//
// Note that it doesn't log until you call Debug, Print, Info, Warn, Fatal
// or Panic on the Entry it returns.
//
// This returns the logrus Entry
func WithFields(fields Fields) *logrus.Entry {
	return logrus.WithFields(logrus.Fields(fields))
}
