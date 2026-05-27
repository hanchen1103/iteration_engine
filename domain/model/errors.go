package model

import "fmt"

type ErrorCode string

const (
	ErrorCodeInvalid   ErrorCode = "INVALID"
	ErrorCodeNotFound  ErrorCode = "NOT_FOUND"
	ErrorCodeConflict  ErrorCode = "CONFLICT"
	ErrorCodeFailed    ErrorCode = "FAILED"
	ErrorCodeForbidden ErrorCode = "FORBIDDEN"
)

type Error struct {
	Code    ErrorCode
	Message string
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Code == "" {
		return e.Message
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func NewError(code ErrorCode, message string) error {
	return &Error{Code: code, Message: message}
}
