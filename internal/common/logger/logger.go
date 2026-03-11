// Package logger provides the application's structured logging using phuslu/log.
package logger

import (
	"os"
	"path/filepath"
	"time"

	"github.com/phuslu/log"
)

// Level defines the severity of a log message.
type Level uint8

const (
	// DebugLevel enables verbose diagnostic logs for development and debugging.
	DebugLevel Level = iota
	// InfoLevel logs general operational events.
	InfoLevel
	// WarnLevel logs unexpected but recoverable conditions.
	WarnLevel
	// ErrorLevel logs failures that require attention.
	ErrorLevel
)

// FieldValue is the set of value types supported for structured log fields.
type FieldValue interface {
	string | bool |
		int | int8 | int16 | int32 | int64 |
		uint | uint8 | uint16 | uint32 | uint64 |
		float32 | float64 |
		time.Time | time.Duration
}

// Field represents a single structured key-value attribute attached to a log event.
type Field struct {
	// Key is the name of the field.
	Key string
	// Value is the field's value.
	Value any
}

// F creates a structured log field from a key and a supported value type.
//
// Example:
//
//	logger.Info("user logged in", logger.F("user_id", 123))
func F[T FieldValue](key string, value T) Field {
	return Field{Key: key, Value: value}
}

// Environment identifies the runtime environment for logger configuration.
type Environment string

const (
	// Development configures logging behavior for local and non-production environments.
	Development Environment = "development"
	// Production configures logging behavior for production environments.
	Production Environment = "production"
)

// Config holds logger initialization settings.
type Config struct {
	// Env specifies the runtime environment (development or production).
	Env Environment
	// Level sets the minimum log level to output.
	Level Level
	// FilePath specifies the log file path for production environments.
	// If empty, logs are written to stderr.
	FilePath string
}

// Configure initializes the logger based on the provided configuration.
//
// In development mode, logs are written to stderr with colored console output.
// In production mode, logs are written to the specified file with rotation.
//
// Example:
//
//	logger.Configure(logger.Config{
//	    Env:   logger.Development,
//	    Level: logger.InfoLevel,
//	})
func Configure(cfg Config) {
	var writer log.Writer

	if cfg.Env == Production && cfg.FilePath != "" {
		dir := filepath.Dir(cfg.FilePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Error().Str("dir", dir).Err(err).Msg("failed to create log directory")
		}
		writer = &log.FileWriter{
			Filename:   cfg.FilePath,
			MaxSize:    100 * 1024 * 1024,
			MaxBackups: 7,
		}
	} else {
		writer = &log.ConsoleWriter{
			ColorOutput: true,
		}
	}

	log.DefaultLogger = log.Logger{
		Level:  log.Level(cfg.Level),
		Writer: writer,
	}
}

// Debug logs a message at debug level.
//
// Example:
//
//	logger.Debug("processing request", logger.F("request_id", "abc123"))
func Debug(msg string, fields ...Field) {
	entry := log.Debug()
	for _, f := range fields {
		entry = addField(entry, f.Key, f.Value)
	}
	entry.Msg(msg)
}

// Info logs a message at info level.
//
// Example:
//
//	logger.Info("user created", logger.F("user_id", 123))
func Info(msg string, fields ...Field) {
	entry := log.Info()
	for _, f := range fields {
		entry = addField(entry, f.Key, f.Value)
	}
	entry.Msg(msg)
}

// Warn logs a message at warning level.
//
// Example:
//
//	logger.Warn("rate limit approaching", logger.F("remaining", 10))
func Warn(msg string, fields ...Field) {
	entry := log.Warn()
	for _, f := range fields {
		entry = addField(entry, f.Key, f.Value)
	}
	entry.Msg(msg)
}

// Error logs a message at error level.
//
// Example:
//
//	logger.Error("database connection failed", logger.F("error", err.Error()))
func Error(msg string, fields ...Field) {
	entry := log.Error()
	for _, f := range fields {
		entry = addField(entry, f.Key, f.Value)
	}
	entry.Msg(msg)
}

func addField(entry *log.Entry, key string, value any) *log.Entry {
	switch v := value.(type) {
	case string:
		entry = entry.Str(key, v)
	case bool:
		entry = entry.Bool(key, v)
	case int:
		entry = entry.Int(key, v)
	case int8:
		entry = entry.Int8(key, v)
	case int16:
		entry = entry.Int16(key, v)
	case int32:
		entry = entry.Int32(key, v)
	case int64:
		entry = entry.Int64(key, v)
	case uint:
		entry = entry.Uint(key, v)
	case uint8:
		entry = entry.Uint8(key, v)
	case uint16:
		entry = entry.Uint16(key, v)
	case uint32:
		entry = entry.Uint32(key, v)
	case uint64:
		entry = entry.Uint64(key, v)
	case float32:
		entry = entry.Float32(key, v)
	case float64:
		entry = entry.Float64(key, v)
	case time.Time:
		entry = entry.Time(key, v)
	case time.Duration:
		entry = entry.Any(key, v)
	default:
		entry = entry.Any(key, v)
	}
	return entry
}

// loggerContext holds fields that are automatically included in all log messages.
type loggerContext struct {
	fields []Field
}

// WithFields returns a context that includes the specified fields in all subsequent log calls.
//
// This is useful for adding request-scoped context like request_id or user_id.
//
// Example:
//
//	ctx := logger.WithFields(logger.F("request_id", "abc123"))
//	ctx.Info("request started")
//	ctx.Info("request completed")
func WithFields(fields ...Field) *loggerContext {
	return &loggerContext{
		fields: fields,
	}
}

func (c *loggerContext) Debug(msg string, fields ...Field) {
	Debug(msg, append(c.fields, fields...)...)
}

func (c *loggerContext) Info(msg string, fields ...Field) {
	Info(msg, append(c.fields, fields...)...)
}

func (c *loggerContext) Warn(msg string, fields ...Field) {
	Warn(msg, append(c.fields, fields...)...)
}

func (c *loggerContext) Error(msg string, fields ...Field) {
	Error(msg, append(c.fields, fields...)...)
}
