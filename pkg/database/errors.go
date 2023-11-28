package database

import "errors"

// ErrorResponse is the form used for Database responses from failures in the DB
type ErrorResponse struct {
	Error  string            `json:"error"`
	Fields map[string]string `json:"fields,omitempty"`
}

// Error is used to pass an error during the request through the
// application with database specific context.
type Error struct {
	Err    error
	Status int
}

// NewError wraps a provided error with an HTTP status code. This
// function should be used when handlers encounter expected errors.
func NewError(err error, status int) error {
	return &Error{
		Err:    err,
		Status: status,
	}
}

// Error implements the error interface. It uses the default message of the
// wrapped error. This is what will be shown in the services' logs.
func (re *Error) Error() string {
	return re.Err.Error()
}

// IsError checks if the error type Error Exists
func IsError(err error) bool {
	var re *Error
	return errors.As(err, &re)
}

// GetError returns a copy of the Error pointer.
func GetError(err error) *Error {
	var re *Error
	if !errors.As(err, &re) {
		return nil
	}
	return re
}
