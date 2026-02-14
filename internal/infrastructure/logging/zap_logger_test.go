package logging

import "testing"

func TestNewLogger_AcceptsTraceLevel(t *testing.T) {
	logger, err := NewLogger(Config{
		Environment: "development",
		Level:       "trace",
	})
	if err != nil {
		t.Fatalf("expected trace level to be accepted, got error: %v", err)
	}
	if logger == nil {
		t.Fatalf("expected logger instance")
	}
}

func TestIsTraceLevel(t *testing.T) {
	if !IsTraceLevel("trace") {
		t.Fatalf("expected trace to be true")
	}
	if !IsTraceLevel(" TRACE ") {
		t.Fatalf("expected TRACE with spaces to be true")
	}
	if IsTraceLevel("debug") {
		t.Fatalf("expected debug to be false")
	}
}
