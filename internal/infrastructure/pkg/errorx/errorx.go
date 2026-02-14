package errorx

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

type Detail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type Error struct {
	code    string
	message string
	kind    Kind
	cause   error
	details []Detail
}

func New(code, message string, kind Kind) *Error {
	return &Error{code: code, message: message, kind: kind}
}

func NewWithDetails(code, message string, kind Kind, details ...Detail) *Error {
	return &Error{code: code, message: message, kind: kind, details: details}
}

func Wrap(code string, kind Kind, err error) *Error {
	if err == nil {
		return nil
	}
	return &Error{code: code, message: err.Error(), kind: kind, cause: err}
}

func InvalidArgument(code, message string, details ...Detail) *Error {
	return NewWithDetails(code, message, KindInvalidArgument, details...)
}

func NotFound(code, message string) *Error {
	return New(code, message, KindNotFound)
}

func Conflict(code, message string) *Error {
	return New(code, message, KindConflict)
}

func Unauthenticated(code, message string) *Error {
	return New(code, message, KindUnauthenticated)
}

func Forbidden(code, message string) *Error {
	return New(code, message, KindForbidden)
}

func Unavailable(code, message string) *Error {
	return New(code, message, KindUnavailable)
}

func Internal(code string, err error) *Error {
	return Wrap(code, KindInternal, err)
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.message != "" {
		return e.message
	}
	if e.cause != nil {
		return e.cause.Error()
	}
	return e.code
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.cause
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

func (e *Error) Details() []Detail {
	if e == nil {
		return nil
	}
	return e.details
}

func Code(err error) string {
	var ce interface {
		Code() string
	}
	if errors.As(err, &ce) {
		return ce.Code()
	}
	return CodeInternal
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

func Details(err error) []Detail {
	var de interface {
		Details() []Detail
	}
	if errors.As(err, &de) {
		return de.Details()
	}
	return nil
}
