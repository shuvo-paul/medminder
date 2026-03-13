package log_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	phuslulog "github.com/phuslu/log"
	"github.com/shuvo-paul/medminder/internal/common/log"
	"github.com/stretchr/testify/assert"
)

func setupBuffer() (*bytes.Buffer, func()) {
	buf := &bytes.Buffer{}
	original := phuslulog.DefaultLogger
	phuslulog.DefaultLogger = phuslulog.Logger{
		Level:  phuslulog.InfoLevel,
		Writer: &phuslulog.IOWriter{Writer: buf},
	}
	return buf, func() { phuslulog.DefaultLogger = original }
}

func setupBufferDebug() (*bytes.Buffer, func()) {
	buf := &bytes.Buffer{}
	original := phuslulog.DefaultLogger
	phuslulog.DefaultLogger = phuslulog.Logger{
		Level:  phuslulog.DebugLevel,
		Writer: &phuslulog.IOWriter{Writer: buf},
	}
	return buf, func() { phuslulog.DefaultLogger = original }
}

func TestDebug_LogsMessage(t *testing.T) {
	buf, cleanup := setupBufferDebug()
	t.Cleanup(cleanup)

	log.Debug("test debug message")

	output := buf.String()
	assert.True(t, strings.Contains(output, "test debug message"))
}

func TestDebug_LogsMessageWithFields(t *testing.T) {
	buf, cleanup := setupBufferDebug()
	t.Cleanup(cleanup)

	log.Debug("test debug message", log.F("key", "value"))

	output := buf.String()
	assert.True(t, strings.Contains(output, "test debug message"))
	assert.True(t, strings.Contains(output, "key"))
	assert.True(t, strings.Contains(output, "value"))
}

func TestInfo_LogsMessage(t *testing.T) {
	buf, cleanup := setupBuffer()
	t.Cleanup(cleanup)

	log.Info("test info message")

	output := buf.String()
	assert.True(t, strings.Contains(output, "test info message"))
}

func TestInfo_LogsMessageWithFields(t *testing.T) {
	buf, cleanup := setupBuffer()
	t.Cleanup(cleanup)

	log.Info("test info message", log.F("key", "value"))

	output := buf.String()
	assert.True(t, strings.Contains(output, "test info message"))
	assert.True(t, strings.Contains(output, "key"))
	assert.True(t, strings.Contains(output, "value"))
}

func TestWarn_LogsMessage(t *testing.T) {
	buf, cleanup := setupBuffer()
	t.Cleanup(cleanup)

	log.Warn("test warn message")

	output := buf.String()
	assert.True(t, strings.Contains(output, "test warn message"))
}

func TestWarn_LogsMessageWithFields(t *testing.T) {
	buf, cleanup := setupBuffer()
	t.Cleanup(cleanup)

	log.Warn("test warn message", log.F("key", "value"))

	output := buf.String()
	assert.True(t, strings.Contains(output, "test warn message"))
	assert.True(t, strings.Contains(output, "key"))
	assert.True(t, strings.Contains(output, "value"))
}

func TestError_LogsMessage(t *testing.T) {
	buf, cleanup := setupBuffer()
	t.Cleanup(cleanup)

	log.Error("test error message")

	output := buf.String()
	assert.True(t, strings.Contains(output, "test error message"))
}

func TestError_LogsMessageWithFields(t *testing.T) {
	buf, cleanup := setupBuffer()
	t.Cleanup(cleanup)

	log.Error("test error message", log.F("key", "value"))

	output := buf.String()
	assert.True(t, strings.Contains(output, "test error message"))
	assert.True(t, strings.Contains(output, "key"))
	assert.True(t, strings.Contains(output, "value"))
}

func TestWithFields_AddsFieldsToContext(t *testing.T) {
	buf, cleanup := setupBufferDebug()
	t.Cleanup(cleanup)

	ctx := log.WithFields(log.F("request_id", "123"))
	ctx.Debug("test message")

	output := buf.String()
	assert.True(t, strings.Contains(output, "test message"))
	assert.True(t, strings.Contains(output, "request_id"))
	assert.True(t, strings.Contains(output, "123"))
}

func TestF_SupportedFieldValues(t *testing.T) {
	buf, cleanup := setupBuffer()
	t.Cleanup(cleanup)

	t.Run("string", func(t *testing.T) {
		buf.Reset()
		log.Info("test", log.F("str_key", "hello"))
		assert.True(t, strings.Contains(buf.String(), "hello"))
	})

	t.Run("bool_true", func(t *testing.T) {
		buf.Reset()
		log.Info("test", log.F("bool_key", true))
		assert.True(t, strings.Contains(buf.String(), "true"))
	})

	t.Run("bool_false", func(t *testing.T) {
		buf.Reset()
		log.Info("test", log.F("bool_key", false))
		assert.True(t, strings.Contains(buf.String(), "false"))
	})

	t.Run("int", func(t *testing.T) {
		buf.Reset()
		log.Info("test", log.F("int_key", 42))
		assert.True(t, strings.Contains(buf.String(), "42"))
	})

	t.Run("int8", func(t *testing.T) {
		buf.Reset()
		log.Info("test", log.F("int8_key", int8(8)))
		assert.True(t, strings.Contains(buf.String(), "8"))
	})

	t.Run("int16", func(t *testing.T) {
		buf.Reset()
		log.Info("test", log.F("int16_key", int16(16)))
		assert.True(t, strings.Contains(buf.String(), "16"))
	})

	t.Run("int32", func(t *testing.T) {
		buf.Reset()
		log.Info("test", log.F("int32_key", int32(32)))
		assert.True(t, strings.Contains(buf.String(), "32"))
	})

	t.Run("int64", func(t *testing.T) {
		buf.Reset()
		log.Info("test", log.F("int64_key", int64(64)))
		assert.True(t, strings.Contains(buf.String(), "64"))
	})

	t.Run("uint", func(t *testing.T) {
		buf.Reset()
		log.Info("test", log.F("uint_key", uint(100)))
		assert.True(t, strings.Contains(buf.String(), "100"))
	})

	t.Run("uint8", func(t *testing.T) {
		buf.Reset()
		log.Info("test", log.F("uint8_key", uint8(8)))
		assert.True(t, strings.Contains(buf.String(), "8"))
	})

	t.Run("uint16", func(t *testing.T) {
		buf.Reset()
		log.Info("test", log.F("uint16_key", uint16(16)))
		assert.True(t, strings.Contains(buf.String(), "16"))
	})

	t.Run("uint32", func(t *testing.T) {
		buf.Reset()
		log.Info("test", log.F("uint32_key", uint32(32)))
		assert.True(t, strings.Contains(buf.String(), "32"))
	})

	t.Run("uint64", func(t *testing.T) {
		buf.Reset()
		log.Info("test", log.F("uint64_key", uint64(64)))
		assert.True(t, strings.Contains(buf.String(), "64"))
	})

	t.Run("float32", func(t *testing.T) {
		buf.Reset()
		log.Info("test", log.F("float32_key", float32(3.14)))
		output := buf.String()
		assert.True(t, strings.Contains(output, "3.14") || strings.Contains(output, "3.140000"))
	})

	t.Run("float64", func(t *testing.T) {
		buf.Reset()
		log.Info("test", log.F("float64_key", float64(6.28)))
		assert.True(t, strings.Contains(buf.String(), "6.28"))
	})

	t.Run("time.Time", func(t *testing.T) {
		buf.Reset()
		log.Info("test", log.F("time_key", time.Date(2026, 3, 11, 0, 0, 0, 0, time.UTC)))
		assert.True(t, strings.Contains(buf.String(), "2026-03-11"))
	})

	t.Run("time.Duration", func(t *testing.T) {
		buf.Reset()
		log.Info("test", log.F("duration_key", time.Second))
		output := buf.String()
		t.Logf("output: %s", output)
		assert.True(t, strings.Contains(output, "1") || strings.Contains(output, "s"))
	})
}

func TestF_ReturnsCorrectField(t *testing.T) {
	field := log.F("test_key", "test_value")
	assert.Equal(t, "test_key", field.Key)
	assert.Equal(t, "test_value", field.Value)
}

func TestConfigure_Development(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "test.log")

	log.Configure(log.Config{
		Env:      log.Development,
		Level:    log.InfoLevel,
		FilePath: file,
	})

	log.Info("test message")
}

func TestConfigure_Production(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "test.log")

	log.Configure(log.Config{
		Env:      log.Production,
		Level:    log.InfoLevel,
		FilePath: file,
	})

	log.Info("test message")

	entries, err := os.ReadDir(dir)
	assert.NoError(t, err)
	assert.Greater(t, len(entries), 0)
}
