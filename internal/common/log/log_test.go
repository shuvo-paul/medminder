package log_test

import (
	"bytes"
	"strings"
	"testing"

	phuslulog "github.com/phuslu/log"
	"github.com/shuvo-paul/medminder/internal/common/log"
	"github.com/stretchr/testify/assert"
)

func TestDebug_LogsMessage(t *testing.T) {
	buf := &bytes.Buffer{}
	phuslulog.DefaultLogger.Writer = &phuslulog.IOWriter{Writer: buf}

	log.Debug("test debug message")

	output := buf.String()
	assert.True(t, strings.Contains(output, "test debug message"))
}

func TestDebug_LogsMessageWithFields(t *testing.T) {
	buf := &bytes.Buffer{}
	phuslulog.DefaultLogger.Writer = &phuslulog.IOWriter{Writer: buf}

	log.Debug("test debug message", log.F("key", "value"))

	output := buf.String()
	assert.True(t, strings.Contains(output, "test debug message"))
	assert.True(t, strings.Contains(output, "key"))
	assert.True(t, strings.Contains(output, "value"))
}

func TestInfo_LogsMessage(t *testing.T) {
	buf := &bytes.Buffer{}
	phuslulog.DefaultLogger.Writer = &phuslulog.IOWriter{Writer: buf}

	log.Info("test info message")

	output := buf.String()
	assert.True(t, strings.Contains(output, "test info message"))
}

func TestInfo_LogsMessageWithFields(t *testing.T) {
	buf := &bytes.Buffer{}
	phuslulog.DefaultLogger.Writer = &phuslulog.IOWriter{Writer: buf}

	log.Info("test info message", log.F("key", "value"))

	output := buf.String()
	assert.True(t, strings.Contains(output, "test info message"))
	assert.True(t, strings.Contains(output, "key"))
	assert.True(t, strings.Contains(output, "value"))
}

func TestWarn_LogsMessage(t *testing.T) {
	buf := &bytes.Buffer{}
	phuslulog.DefaultLogger.Writer = &phuslulog.IOWriter{Writer: buf}

	log.Warn("test warn message")

	output := buf.String()
	assert.True(t, strings.Contains(output, "test warn message"))
}

func TestWarn_LogsMessageWithFields(t *testing.T) {
	buf := &bytes.Buffer{}
	phuslulog.DefaultLogger.Writer = &phuslulog.IOWriter{Writer: buf}

	log.Warn("test warn message", log.F("key", "value"))

	output := buf.String()
	assert.True(t, strings.Contains(output, "test warn message"))
	assert.True(t, strings.Contains(output, "key"))
	assert.True(t, strings.Contains(output, "value"))
}

func TestError_LogsMessage(t *testing.T) {
	buf := &bytes.Buffer{}
	phuslulog.DefaultLogger.Writer = &phuslulog.IOWriter{Writer: buf}

	log.Error("test error message")

	output := buf.String()
	assert.True(t, strings.Contains(output, "test error message"))
}

func TestError_LogsMessageWithFields(t *testing.T) {
	buf := &bytes.Buffer{}
	phuslulog.DefaultLogger.Writer = &phuslulog.IOWriter{Writer: buf}

	log.Error("test error message", log.F("key", "value"))

	output := buf.String()
	assert.True(t, strings.Contains(output, "test error message"))
	assert.True(t, strings.Contains(output, "key"))
	assert.True(t, strings.Contains(output, "value"))
}

func TestWithFields_AddsFieldsToContext(t *testing.T) {
	buf := &bytes.Buffer{}
	phuslulog.DefaultLogger.Writer = &phuslulog.IOWriter{Writer: buf}

	ctx := log.WithFields(log.F("request_id", "123"))
	ctx.Debug("test message")

	output := buf.String()
	assert.True(t, strings.Contains(output, "test message"))
	assert.True(t, strings.Contains(output, "request_id"))
	assert.True(t, strings.Contains(output, "123"))
}
