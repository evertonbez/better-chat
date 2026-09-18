package apperr

import (
	"errors"
	"fmt"
	"maps"
	"net/http"
)

type Code string

const (
	CodeInvalidRequest     Code = "INVALID_REQUEST"
	CodeValidation         Code = "VALIDATION"
	CodeUnauthorized       Code = "UNAUTHORIZED"
	CodeForbidden          Code = "FORBIDDEN"
	CodeInvalidCredentials Code = "INVALID_CREDENTIALS"
	CodeSessionExpired     Code = "SESSION_EXPIRED"
	CodeEmailTaken         Code = "EMAIL_TAKEN"
	CodeEmailNotVerified   Code = "EMAIL_NOT_VERIFIED"
	CodeUserNotFound       Code = "USER_NOT_FOUND"
	CodeAccountNotFound    Code = "ACCOUNT_NOT_FOUND"
	CodeTooManyRequests    Code = "TOO_MANY_REQUESTS"
	CodeInternal           Code = "INTERNAL"
)

type Error struct {
	Code    Code
	Status  int
	Message string

	Details map[string]any
	cause   error
}

func (e *Error) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *Error) Unwrap() error { return e.cause }

func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	return ok && t.Code == e.Code
}

func (e *Error) Msg(format string, args ...any) *Error {
	c := *e
	c.Message = fmt.Sprintf(format, args...)
	return &c
}

func (e *Error) Cause(err error) *Error {
	c := *e
	c.cause = err
	return &c
}

func (e *Error) With(key string, value any) *Error {
	c := *e
	c.Details = maps.Clone(e.Details)
	if c.Details == nil {
		c.Details = make(map[string]any, 1)
	}
	c.Details[key] = value
	return &c
}

var (
	InvalidRequest = &Error{Code: CodeInvalidRequest, Status: http.StatusBadRequest, Message: "malformed request body"}

	Unauthorized       = &Error{Code: CodeUnauthorized, Status: http.StatusUnauthorized, Message: "authentication required"}
	InvalidCredentials = &Error{Code: CodeInvalidCredentials, Status: http.StatusUnauthorized, Message: "invalid email or password"}
	SessionExpired     = &Error{Code: CodeSessionExpired, Status: http.StatusUnauthorized, Message: "session expired"}

	Forbidden        = &Error{Code: CodeForbidden, Status: http.StatusForbidden, Message: "forbidden"}
	EmailNotVerified = &Error{Code: CodeEmailNotVerified, Status: http.StatusForbidden, Message: "email not verified"}

	UserNotFound    = &Error{Code: CodeUserNotFound, Status: http.StatusNotFound, Message: "user not found"}
	AccountNotFound = &Error{Code: CodeAccountNotFound, Status: http.StatusNotFound, Message: "account not found"}

	EmailTaken = &Error{Code: CodeEmailTaken, Status: http.StatusConflict, Message: "email already registered"}

	Validation = &Error{Code: CodeValidation, Status: http.StatusUnprocessableEntity, Message: "invalid input"}

	TooManyRequests = &Error{Code: CodeTooManyRequests, Status: http.StatusTooManyRequests, Message: "too many requests"}

	Internal = &Error{Code: CodeInternal, Status: http.StatusInternalServerError, Message: "internal error"}
)

func From(err error) *Error {
	if err == nil {
		return nil
	}

	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr
	}

	return Internal.Cause(err)
}
