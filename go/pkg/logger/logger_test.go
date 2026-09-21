package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
)

func TestNewJSONLoggerWritesAttrs(t *testing.T) {
	var buf bytes.Buffer

	log, err := New(Options{
		Level:  "debug",
		Format: FormatJSON,
		Output: &buf,
		Attrs:  []slog.Attr{slog.String("service", "api")},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	log.Debug("hello", "count", 1)

	var record map[string]any
	if err := json.Unmarshal(buf.Bytes(), &record); err != nil {
		t.Fatalf("decode record %q: %v", buf.String(), err)
	}

	if record["msg"] != "hello" {
		t.Errorf("msg = %v, want %q", record["msg"], "hello")
	}
	if record["service"] != "api" {
		t.Errorf("service = %v, want %q", record["service"], "api")
	}
	if record["level"] != "DEBUG" {
		t.Errorf("level = %v, want %q", record["level"], "DEBUG")
	}
}

func TestNewRespectsLevel(t *testing.T) {
	var buf bytes.Buffer

	log, err := New(Options{Level: "warn", Output: &buf})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	log.Info("dropped")
	if buf.Len() != 0 {
		t.Fatalf("info record was written at warn level: %q", buf.String())
	}

	log.Warn("kept")
	if !strings.Contains(buf.String(), "kept") {
		t.Fatalf("warn record missing: %q", buf.String())
	}
}

func TestNewRejectsUnknownFormatAndLevel(t *testing.T) {
	if _, err := New(Options{Format: "xml"}); err == nil {
		t.Error("New with unknown format: want error, got nil")
	}

	if _, err := New(Options{Level: "loud"}); err == nil {
		t.Error("New with unknown level: want error, got nil")
	}
}

func TestParseLevelDefaultsToInfo(t *testing.T) {
	level, err := ParseLevel("")
	if err != nil {
		t.Fatalf("ParseLevel: %v", err)
	}
	if level != slog.LevelInfo {
		t.Errorf("ParseLevel(\"\") = %v, want %v", level, slog.LevelInfo)
	}
}

func TestContextRoundTrip(t *testing.T) {
	log, err := New(Options{Output: &bytes.Buffer{}})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ctx := WithContext(context.Background(), log)
	if got := FromContext(ctx); got != log {
		t.Error("FromContext did not return the stored logger")
	}

	if got := FromContext(context.Background()); got != slog.Default() {
		t.Error("FromContext without a logger should fall back to the default")
	}
}
