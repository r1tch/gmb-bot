package errors

import "fmt"

type Kind string

const (
	KindConfig     Kind = "config"
	KindExtraction Kind = "extraction"
	KindTransport  Kind = "transport"
	KindState      Kind = "state"
	KindTransient  Kind = "transient"
	KindPermanent  Kind = "permanent"
)

type AppError struct {
	Kind Kind
	Op   string
	Err  error
}

func (e *AppError) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Op == "" {
		return fmt.Sprintf("%s: %v", e.Kind, e.Err)
	}
	return fmt.Sprintf("%s %s: %v", e.Kind, e.Op, e.Err)
}

func (e *AppError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func Wrap(kind Kind, op string, err error) error {
	if err == nil {
		return nil
	}
	return &AppError{Kind: kind, Op: op, Err: err}
}
