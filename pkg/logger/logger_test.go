package logger

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/rs/zerolog"
)

func TestLogger(t *testing.T) {
	var buf bytes.Buffer

	log := &ZeroLogger{
		log: zerolog.New(&buf).With().Timestamp().Logger(),
	}

	testCases := []struct {
		name   string
		logFn  func(msg string, fields map[string]interface{})
		level  string
		msg    string
		fields map[string]interface{}
	}{
		{
			name:   "info log",
			logFn:  log.Info,
			level:  "info",
			msg:    "test info",
			fields: map[string]interface{}{"key": "value"},
		},
		{
			name:   "debug log",
			logFn:  log.Debug,
			level:  "debug",
			msg:    "test debug",
			fields: map[string]interface{}{"number": float64(42)}, // Изменили тип на float64
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf.Reset()
			tc.logFn(tc.msg, tc.fields)

			var logEntry map[string]interface{}
			if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
				t.Fatalf("Failed to unmarshal log entry: %v", err)
			}

			if logEntry["message"] != tc.msg {
				t.Errorf("Expected message %q, got %q", tc.msg, logEntry["message"])
			}

			if logEntry["level"] != tc.level {
				t.Errorf("Expected level %q, got %q", tc.level, logEntry["level"])
			}

			for k, v := range tc.fields {
				if logEntry[k] != v {
					t.Errorf("Expected field %q to be %v (%T), got %v (%T)",
						k, v, v, logEntry[k], logEntry[k])
				}
			}
		})
	}
}

func TestWithComponent(t *testing.T) {
	var buf bytes.Buffer
	log := &ZeroLogger{
		log: zerolog.New(&buf).With().Timestamp().Logger(),
	}

	componentName := "test-component"
	componentLog := log.WithComponent(componentName)

	componentLog.Info("test message", nil)

	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to unmarshal log entry: %v", err)
	}

	if logEntry["component"] != componentName {
		t.Errorf("Expected component %q, got %q", componentName, logEntry["component"])
	}
}
