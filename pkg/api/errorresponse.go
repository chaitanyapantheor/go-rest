package api

import (
	"errors"
)

// ErrorResponse is the form used for API responses from failures in the API.
// swagger:model ErrorResponse
type ErrorResponse struct {
	// in:body
	//
	//example: data is not in proper format
	Error string `json:"error"`
	// in:body
	//
	//example: {"field": "error message for this specific field"}
	Fields map[string]string `json:"fields,omitempty"`
}

// ErrorResponseID is the form used for API responses from failures in the API.
// swagger:model ErrorResponseID
type ErrorResponseID struct {
	// in:body
	//
	//example: ID is not in its proper form
	Error string `json:"error"`
	// in:body
	//
	//example: {"field": "error message for this specific field"}
	Fields map[string]string `json:"fields,omitempty"`
}

// ErrorResponseIDs is the form used for API responses from failures in the API.
// swagger:model ErrorResponseIDs
type ErrorResponseIDs struct {
	// in:body
	//
	//example: IDs are not in their proper form
	Error string `json:"error"`
	// in:body
	//
	//example: {"field": "error message for this specific field"}
	Fields map[string]string `json:"fields,omitempty"`
}

// ErrorResponseUUID is the form used for API responses from failures in the API.
// swagger:model ErrorResponseUUID
type ErrorResponseUUID struct {
	// in:body
	//
	//example: UUID is not in its proper form
	Error string `json:"error"`
	// in:body
	//
	//example: {"field": "error message for this specific field"}
	Fields map[string]string `json:"fields,omitempty"`
}

// RequestError is used to pass an error during the request through the
// application with web specific context.
type RequestError struct {
	Err    error
	Status int
}

// NewRequestError wraps a provided error with an HTTP status code. This
// function should be used when handlers encounter expected errors.
func NewRequestError(err error, status int) error {
	return &RequestError{
		Err:    err,
		Status: status,
	}
}

// Error implements the error interface. It uses the default message of the
// wrapped error. This is what will be shown in the services' logs.
func (re *RequestError) Error() string {
	return re.Err.Error()
}

// IsRequestError checks if the error type RequestError Exists
func IsRequestError(err error) bool {
	var re *RequestError
	return errors.As(err, &re)
}

// GetRequestError returns a copy of the RequestError pointer.
func GetRequestError(err error) *RequestError {
	var re *RequestError
	if !errors.As(err, &re) {
		return nil
	}
	return re
}
