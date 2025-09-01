package log

import (
	"os"
	"runtime"

	"github.com/sirupsen/logrus"
)

var Logger *logrus.Logger

// Init initializes the logger with configuration
func Init(level string, format string) {
	Logger = logrus.New()

	// Set log level
	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		logLevel = logrus.InfoLevel
	}
	Logger.SetLevel(logLevel)

	// Set output to stdout
	Logger.SetOutput(os.Stdout)

	// Set formatter based on configuration
	switch format {
	case "json":
		Logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
			CallerPrettyfier: func(f *runtime.Frame) (string, string) {
				filename := f.File
				if len(filename) > 20 {
					filename = "..." + filename[len(filename)-17:]
				}
				return "", filename + ":" + string(rune(f.Line))
			},
		})
	case "text":
		Logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
			DisableColors:   false,
		})
	default:
		Logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
			DisableColors:   false,
		})
	}

	// Enable caller information for debug and trace levels
	if logLevel <= logrus.DebugLevel {
		Logger.SetReportCaller(true)
	}

	// Log initialization
	Logger.WithFields(logrus.Fields{
		"level":  logLevel.String(),
		"format": format,
	}).Info("Logger initialized")
}

// GetLogger returns the configured logger instance
func GetLogger() *logrus.Logger {
	if Logger == nil {
		// Initialize with defaults if not already initialized
		Init("info", "text")
	}
	return Logger
}

// WithField creates a new entry with a single field
func WithField(key string, value interface{}) *logrus.Entry {
	return GetLogger().WithField(key, value)
}

// WithFields creates a new entry with multiple fields
func WithFields(fields logrus.Fields) *logrus.Entry {
	return GetLogger().WithFields(fields)
}

// WithError creates a new entry with error information
func WithError(err error) *logrus.Entry {
	return GetLogger().WithError(err)
}

// Debug logs a debug message
func Debug(args ...interface{}) {
	GetLogger().Debug(args...)
}

// Debugf logs a formatted debug message
func Debugf(format string, args ...interface{}) {
	GetLogger().Debugf(format, args...)
}

// Info logs an info message
func Info(args ...interface{}) {
	GetLogger().Info(args...)
}

// Infof logs a formatted info message
func Infof(format string, args ...interface{}) {
	GetLogger().Infof(format, args...)
}

// Warn logs a warning message
func Warn(args ...interface{}) {
	GetLogger().Warn(args...)
}

// Warnf logs a formatted warning message
func Warnf(format string, args ...interface{}) {
	GetLogger().Warnf(format, args...)
}

// Error logs an error message
func Error(args ...interface{}) {
	GetLogger().Error(args...)
}

// Errorf logs a formatted error message
func Errorf(format string, args ...interface{}) {
	GetLogger().Errorf(format, args...)
}

// Fatal logs a fatal message and exits
func Fatal(args ...interface{}) {
	GetLogger().Fatal(args...)
}

// Fatalf logs a formatted fatal message and exits
func Fatalf(format string, args ...interface{}) {
	GetLogger().Fatalf(format, args...)
}

// Panic logs a panic message and panics
func Panic(args ...interface{}) {
	GetLogger().Panic(args...)
}

// Panicf logs a formatted panic message and panics
func Panicf(format string, args ...interface{}) {
	GetLogger().Panicf(format, args...)
}
