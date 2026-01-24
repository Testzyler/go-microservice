package errx

import "errors"

type Kind int

const (
	KindInternal Kind = iota
	KindInvalidArgument
	KindNotFound
	KindConflict
	KindUnauthenticated
	KindForbidden
	KindUnavailable
)

type Error struct {
	code  string
	msg   string
	kind  Kind
	cause error
}

func New(code, msg string, kind Kind) *Error {
	return &Error{code: code, msg: msg, kind: kind}
}

func Wrap(code string, kind Kind, err error) *Error {
	if err == nil {
		return nil
	}
	return &Error{code: code, msg: err.Error(), kind: kind, cause: err}
}

func (e *Error) WithCause(err error) *Error {
	if e == nil {
		return nil
	}
	return &Error{code: e.code, msg: e.msg, kind: e.kind, cause: err}
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.msg != "" {
		if e.cause != nil && e.msg != e.cause.Error() {
			return e.msg + ": " + e.cause.Error()
		}
		return e.msg
	}
	if e.cause != nil {
		return e.cause.Error()
	}
	return e.code
}

func (e *Error) Code() string {
	if e == nil {
		return ""
	}
	return e.code
}

func (e *Error) Kind() Kind {
	if e == nil {
		return KindInternal
	}
	return e.kind
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.cause
}

func Code(err error) string {
	var ce interface {
		Code() string
	}
	if errors.As(err, &ce) {
		return ce.Code()
	}
	return "internal"
}

func KindOf(err error) Kind {
	var ke interface {
		Kind() Kind
	}
	if errors.As(err, &ke) {
		return ke.Kind()
	}
	return KindInternal
}
