package main

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestNewLogger(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&buf, "info")

	if logger == nil {
		t.Fatal("NewLogger() 返回 nil")
	}
	if logger.out != &buf {
		t.Error("NewLogger() out 未正确设置")
	}
	if logger.level != "info" {
		t.Errorf("NewLogger() level = %q, want %q", logger.level, "info")
	}
}

func TestNewStdoutLogger(t *testing.T) {
	logger := NewStdoutLogger("debug")

	if logger == nil {
		t.Fatal("NewStdoutLogger() 返回 nil")
	}
	if logger.level != "debug" {
		t.Errorf("NewStdoutLogger() level = %q, want %q", logger.level, "debug")
	}
}

func TestLoggerWithFields(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&buf, "info")

	fields := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
	}

	newLogger := logger.WithFields(fields)

	if newLogger == logger {
		t.Error("WithFields() 应该返回新的 Logger 实例")
	}
	if len(newLogger.fields) != 2 {
		t.Errorf("WithFields() fields 长度 = %d, want 2", len(newLogger.fields))
	}
	if newLogger.fields["key1"] != "value1" {
		t.Errorf("WithFields() fields[key1] = %v, want %q", newLogger.fields["key1"], "value1")
	}
	if newLogger.fields["key2"] != 42 {
		t.Errorf("WithFields() fields[key2] = %v, want 42", newLogger.fields["key2"])
	}
}

func TestLoggerWithFieldsMerge(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&buf, "info")

	logger = logger.WithFields(map[string]interface{}{
		"key1": "original",
	})

	newLogger := logger.WithFields(map[string]interface{}{
		"key1": "overridden",
		"key2": "new",
	})

	if newLogger.fields["key1"] != "overridden" {
		t.Errorf("WithFields() 应该覆盖已有的 key1, got %v", newLogger.fields["key1"])
	}
	if newLogger.fields["key2"] != "new" {
		t.Errorf("WithFields() 应该添加新的 key2, got %v", newLogger.fields["key2"])
	}
}

func TestLoggerLevels(t *testing.T) {
	tests := []struct {
		name     string
		level    string
		logLevel string
		shouldLog bool
	}{
		{"debug level logs debug", "debug", "debug", true},
		{"info level logs info", "info", "info", true},
		{"info level does not log debug", "info", "debug", false},
		{"warn level logs warn", "warn", "warn", true},
		{"warn level does not log info", "warn", "info", false},
		{"error level logs error", "error", "error", true},
		{"error level does not log warn", "error", "warn", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := NewLogger(&buf, tt.level)
			logger.log(tt.logLevel, "test message")

			if tt.shouldLog && buf.Len() == 0 {
				t.Errorf("level %q 应该记录 %q 级别的日志", tt.level, tt.logLevel)
			}
			if !tt.shouldLog && buf.Len() > 0 {
				t.Errorf("level %q 不应该记录 %q 级别的日志", tt.level, tt.logLevel)
			}
		})
	}
}

func TestLoggerOutput(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&buf, "info")
	logger.Info("test message")

	output := buf.String()
	if output == "" {
		t.Fatal("Logger.Info() 没有输出")
	}

	var entry LogEntry
	if err := json.Unmarshal([]byte(output), &entry); err != nil {
		t.Fatalf("Logger.Info() 输出不是有效的 JSON: %v", err)
	}

	if entry.Level != "info" {
		t.Errorf("entry.Level = %q, want %q", entry.Level, "info")
	}
	if entry.Message != "test message" {
		t.Errorf("entry.Message = %q, want %q", entry.Message, "test message")
	}
	if entry.Time == "" {
		t.Error("entry.Time 不应为空")
	}
}

func TestLoggerFormattedOutput(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&buf, "info")
	logger.Infof("hello %s, you are %d years old", "world", 42)

	var entry LogEntry
	if err := json.Unmarshal([]byte(buf.String()), &entry); err != nil {
		t.Fatalf("Logger.Infof() 输出不是有效的 JSON: %v", err)
	}

	expected := "hello world, you are 42 years old"
	if entry.Message != expected {
		t.Errorf("entry.Message = %q, want %q", entry.Message, expected)
	}
}

func TestLoggerWithFieldsOutput(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&buf, "info")
	logger = logger.WithFields(map[string]interface{}{
		"request_id": "12345",
	})
	logger.Info("request processed")

	var entry LogEntry
	if err := json.Unmarshal([]byte(buf.String()), &entry); err != nil {
		t.Fatalf("Logger.WithFields().Info() 输出不是有效的 JSON: %v", err)
	}

	if entry.Fields == nil {
		t.Fatal("entry.Fields 不应为 nil")
	}
	if entry.Fields["request_id"] != "12345" {
		t.Errorf("entry.Fields[request_id] = %v, want %q", entry.Fields["request_id"], "12345")
	}
}

func TestLoggerDebug(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&buf, "debug")
	logger.Debug("debug message")

	var entry LogEntry
	if err := json.Unmarshal([]byte(buf.String()), &entry); err != nil {
		t.Fatalf("Logger.Debug() 输出不是有效的 JSON: %v", err)
	}

	if entry.Level != "debug" {
		t.Errorf("entry.Level = %q, want %q", entry.Level, "debug")
	}
	if entry.Message != "debug message" {
		t.Errorf("entry.Message = %q, want %q", entry.Message, "debug message")
	}
}

func TestLoggerWarn(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&buf, "warn")
	logger.Warn("warning message")

	var entry LogEntry
	if err := json.Unmarshal([]byte(buf.String()), &entry); err != nil {
		t.Fatalf("Logger.Warn() 输出不是有效的 JSON: %v", err)
	}

	if entry.Level != "warn" {
		t.Errorf("entry.Level = %q, want %q", entry.Level, "warn")
	}
	if entry.Message != "warning message" {
		t.Errorf("entry.Message = %q, want %q", entry.Message, "warning message")
	}
}

func TestLoggerError(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&buf, "error")
	logger.Error("error message")

	var entry LogEntry
	if err := json.Unmarshal([]byte(buf.String()), &entry); err != nil {
		t.Fatalf("Logger.Error() 输出不是有效的 JSON: %v", err)
	}

	if entry.Level != "error" {
		t.Errorf("entry.Level = %q, want %q", entry.Level, "error")
	}
	if entry.Message != "error message" {
		t.Errorf("entry.Message = %q, want %q", entry.Message, "error message")
	}
}

func TestLoggerDefaultLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&buf, "")

	logger.log("info", "test")

	if buf.Len() == 0 {
		t.Error("空 level 应该默认为 info 并记录日志")
	}
}

func TestLoggerDebugf(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&buf, "debug")
	logger.Debugf("debug %s %d", "message", 42)

	var entry LogEntry
	if err := json.Unmarshal([]byte(buf.String()), &entry); err != nil {
		t.Fatalf("Logger.Debugf() 输出不是有效的 JSON: %v", err)
	}

	if entry.Level != "debug" {
		t.Errorf("entry.Level = %q, want %q", entry.Level, "debug")
	}
	if entry.Message != "debug message 42" {
		t.Errorf("entry.Message = %q, want %q", entry.Message, "debug message 42")
	}
}

func TestLoggerWarnf(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&buf, "warn")
	logger.Warnf("warn %s %d", "message", 42)

	var entry LogEntry
	if err := json.Unmarshal([]byte(buf.String()), &entry); err != nil {
		t.Fatalf("Logger.Warnf() 输出不是有效的 JSON: %v", err)
	}

	if entry.Level != "warn" {
		t.Errorf("entry.Level = %q, want %q", entry.Level, "warn")
	}
	if entry.Message != "warn message 42" {
		t.Errorf("entry.Message = %q, want %q", entry.Message, "warn message 42")
	}
}

func TestLoggerErrorf(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&buf, "error")
	logger.Errorf("error %s %d", "message", 42)

	var entry LogEntry
	if err := json.Unmarshal([]byte(buf.String()), &entry); err != nil {
		t.Fatalf("Logger.Errorf() 输出不是有效的 JSON: %v", err)
	}

	if entry.Level != "error" {
		t.Errorf("entry.Level = %q, want %q", entry.Level, "error")
	}
	if entry.Message != "error message 42" {
		t.Errorf("entry.Message = %q, want %q", entry.Message, "error message 42")
	}
}
