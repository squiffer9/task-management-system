package logger

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"time"
)

// Level represents the severity level of a log message
type Level int

const (
	// LevelDebug level for verbose messages useful for debugging
	LevelDebug Level = iota
	// LevelInfo level for general operational information
	LevelInfo
	// LevelWarn level for non-critical issues that might need attention
	LevelWarn
	// LevelError level for errors that should be addressed
	LevelError
	// LevelFatal level for critical errors that lead to termination
	LevelFatal
)

var levelNames = map[Level]string{
	LevelDebug: "DEBUG",
	LevelInfo:  "INFO",
	LevelWarn:  "WARN",
	LevelError: "ERROR",
	LevelFatal: "FATAL",
}

// Logger represents a simple structured logger
type Logger struct {
	level  Level
	writer io.Writer
}

// New creates a new logger instance with the specified minimum level
func New(level Level) *Logger {
	return &Logger{
		level:  level,
		writer: os.Stdout,
	}
}

// SetWriter sets the writer where logs will be written to
func (l *Logger) SetWriter(writer io.Writer) {
	l.writer = writer
}

// SetLevel sets the minimum log level
func (l *Logger) SetLevel(level Level) {
	l.level = level
}

// log writes a log message with the specified level and fields
func (l *Logger) log(level Level, msg string, fields map[string]interface{}) {
	if level < l.level {
		return
	}

	now := time.Now().Format(time.RFC3339)
	levelName := levelNames[level]

	// Get caller information
	_, file, line, ok := runtime.Caller(2)
	caller := "unknown"
	if ok {
		// Extract just the file name without the full path
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				file = file[i+1:]
				break
			}
		}
		caller = fmt.Sprintf("%s:%d", file, line)
	}

	// Format the log message with basic fields
	logEntry := fmt.Sprintf("[%s] [%s] [%s] %s", now, levelName, caller, msg)

	// Add additional fields if present
	if len(fields) > 0 {
		logEntry += " "
		for k, v := range fields {
			logEntry += fmt.Sprintf("%s=%v ", k, v)
		}
	}

	fmt.Fprintln(l.writer, logEntry)

	// For fatal logs, terminate the program
	if level == LevelFatal {
		os.Exit(1)
	}
}

// Debug logs a message at debug level
func (l *Logger) Debug(msg string, fields map[string]interface{}) {
	l.log(LevelDebug, msg, fields)
}

// Info logs a message at info level
func (l *Logger) Info(msg string, fields map[string]interface{}) {
	l.log(LevelInfo, msg, fields)
}

// Warn logs a message at warn level
func (l *Logger) Warn(msg string, fields map[string]interface{}) {
	l.log(LevelWarn, msg, fields)
}

// Error logs a message at error level
func (l *Logger) Error(msg string, fields map[string]interface{}) {
	l.log(LevelError, msg, fields)
}

// Fatal logs a message at fatal level and terminates the program
func (l *Logger) Fatal(msg string, fields map[string]interface{}) {
	l.log(LevelFatal, msg, fields)
}

// DebugF logs a debug message with formatted string
func (l *Logger) DebugF(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.Debug(msg, nil)
}

// InfoF logs an info message with formatted string
func (l *Logger) InfoF(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.Info(msg, nil)
}

// WarnF logs a warning message with formatted string
func (l *Logger) WarnF(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.Warn(msg, nil)
}

// ErrorF logs an error message with formatted string
func (l *Logger) ErrorF(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.Error(msg, nil)
}

// FatalF logs a fatal message with formatted string and terminates the program
func (l *Logger) FatalF(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.Fatal(msg, nil)
}

// Global default logger instance
var defaultLogger = New(LevelInfo)

// SetDefaultLevel sets the log level for the default logger
func SetDefaultLevel(level Level) {
	defaultLogger.SetLevel(level)
}

// SetDefaultWriter sets the writer for the default logger
func SetDefaultWriter(writer io.Writer) {
	defaultLogger.SetWriter(writer)
}

// Global logging functions

// DebugF logs a formatted debug message using the default logger
func DebugF(format string, args ...interface{}) {
	defaultLogger.DebugF(format, args...)
}

// InfoF logs a formatted info message using the default logger
func InfoF(format string, args ...interface{}) {
	defaultLogger.InfoF(format, args...)
}

// WarnF logs a formatted warning message using the default logger
func WarnF(format string, args ...interface{}) {
	defaultLogger.WarnF(format, args...)
}

// ErrorF logs a formatted error message using the default logger
func ErrorF(format string, args ...interface{}) {
	defaultLogger.ErrorF(format, args...)
}

// FatalF logs a formatted fatal message using the default logger and terminates the program
func FatalF(format string, args ...interface{}) {
	defaultLogger.FatalF(format, args...)
}

// Debug logs a message at debug level using the default logger
func Debug(msg string, fields map[string]interface{}) {
	defaultLogger.Debug(msg, fields)
}

// Info logs a message at info level using the default logger
func Info(msg string, fields map[string]interface{}) {
	defaultLogger.Info(msg, fields)
}

// Warn logs a message at warn level using the default logger
func Warn(msg string, fields map[string]interface{}) {
	defaultLogger.Warn(msg, fields)
}

// Error logs a message at error level using the default logger
func Error(msg string, fields map[string]interface{}) {
	defaultLogger.Error(msg, fields)
}

// Fatal logs a message at fatal level using the default logger and terminates the program
func Fatal(msg string, fields map[string]interface{}) {
	defaultLogger.Fatal(msg, fields)
}
