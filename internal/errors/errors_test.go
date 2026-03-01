package errors

import (
	"errors"
	"testing"
)

func TestWrapAndUnwrap(t *testing.T) {
	base := errors.New("boom")
	err := Wrap(KindTransport, "op", base)
	if err == nil {
		t.Fatal("expected non-nil wrapped error")
	}
	appErr, ok := err.(*AppError)
	if !ok {
		t.Fatalf("expected *AppError, got %T", err)
	}
	if appErr.Kind != KindTransport || appErr.Op != "op" {
		t.Fatalf("unexpected app error fields: %+v", appErr)
	}
	if !errors.Is(err, base) {
		t.Fatalf("expected errors.Is to match wrapped base error")
	}
}

func TestWrapNil(t *testing.T) {
	if Wrap(KindConfig, "x", nil) != nil {
		t.Fatal("expected nil when wrapping nil")
	}
}
