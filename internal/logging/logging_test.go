package logging

import "testing"

func TestNewLogger(t *testing.T) {
	for _, lvl := range []string{"debug", "info", "warn", "error", "unknown"} {
		if logger := New(lvl); logger == nil {
			t.Fatalf("expected non-nil logger for level %q", lvl)
		}
	}
}
