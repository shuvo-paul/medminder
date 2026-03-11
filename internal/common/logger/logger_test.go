package logger_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/phuslu/log"
	"github.com/shuvo-paul/medminder/internal/common/logger"
	"github.com/stretchr/testify/assert"
)

func TestDebug_LogsMessage(t *testing.T) {
	buf := &bytes.Buffer{}
	log.DefaultLogger.Writer = &log.IOWriter{Writer: buf}

	logger.Debug("test debug message")

	output := buf.String()
	assert.True(t, strings.Contains(output, "test debug message"))
}

func TestDebug_LogsMessageWithFields(t *testing.T) {
	buf := &bytes.Buffer{}
	log.DefaultLogger.Writer = &log.IOWriter{Writer: buf}

	logger.Debug("test debug message", logger.F("key", "value"))

	output := buf.String()
	assert.True(t, strings.Contains(output, "test debug message"))
	assert.True(t, strings.Contains(output, "key"))
	assert.True(t, strings.Contains(output, "value"))
}

func TestInfo_LogsMessage(t *testing.T) {
	buf := &bytes.Buffer{}
	log.DefaultLogger.Writer = &log.IOWriter{Writer: buf}

	logger.Info("test info message")

	output := buf.String()
	assert.True(t, strings.Contains(output, "test info message"))
}

func TestInfo_LogsMessageWithFields(t *testing.T) {
	buf := &bytes.Buffer{}
	log.DefaultLogger.Writer = &log.IOWriter{Writer: buf}

	logger.Info("test info message", logger.F("key", "value"))

	output := buf.String()
	assert.True(t, strings.Contains(output, "test info message"))
	assert.True(t, strings.Contains(output, "key"))
	assert.True(t, strings.Contains(output, "value"))
}

func TestWarn_LogsMessage(t *testing.T) {
	buf := &bytes.Buffer{}
	log.DefaultLogger.Writer = &log.IOWriter{Writer: buf}

	logger.Warn("test warn message")

	output := buf.String()
	assert.True(t, strings.Contains(output, "test warn message"))
}

func TestWarn_LogsMessageWithFields(t *testing.T) {
	buf := &bytes.Buffer{}
	log.DefaultLogger.Writer = &log.IOWriter{Writer: buf}

	logger.Warn("test warn message", logger.F("key", "value"))

	output := buf.String()
	assert.True(t, strings.Contains(output, "test warn message"))
	assert.True(t, strings.Contains(output, "key"))
	assert.True(t, strings.Contains(output, "value"))
}

func TestError_LogsMessage(t *testing.T) {
	buf := &bytes.Buffer{}
	log.DefaultLogger.Writer = &log.IOWriter{Writer: buf}

	logger.Error("test error message")

	output := buf.String()
	assert.True(t, strings.Contains(output, "test error message"))
}

func TestError_LogsMessageWithFields(t *testing.T) {
	buf := &bytes.Buffer{}
	log.DefaultLogger.Writer = &log.IOWriter{Writer: buf}

	logger.Error("test error message", logger.F("key", "value"))

	output := buf.String()
	assert.True(t, strings.Contains(output, "test error message"))
	assert.True(t, strings.Contains(output, "key"))
	assert.True(t, strings.Contains(output, "value"))
}

func TestWithFields_AddsFieldsToContext(t *testing.T) {
	buf := &bytes.Buffer{}
	log.DefaultLogger.Writer = &log.IOWriter{Writer: buf}

	ctx := logger.WithFields(logger.F("request_id", "123"))
	ctx.Debug("test message")

	output := buf.String()
	assert.True(t, strings.Contains(output, "test message"))
	assert.True(t, strings.Contains(output, "request_id"))
	assert.True(t, strings.Contains(output, "123"))
}
