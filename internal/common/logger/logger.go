// Package logger provides the application's structured logging
package logger

import "time"

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
	Key   string
	Value any
}

// F creates a structured log field from a key and a supported value type.
func F[T FieldValue](key string, value T) Field {
	return Field{Key: key, Value: value}
}

// Logger defines the application logging contract.
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
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
	Env      Environment
	Level    Level
	FilePath string
}
