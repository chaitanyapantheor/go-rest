package api

import (
	"context"
	"errors"
	"time"
)

// ctxKey represents the type of value for the context key.
type ctxKey int

// key is how request values are stored/retrieved.
const key ctxKey = 1

// ContextValues represent state for each request.
type ContextValues struct {
	TracerUID  string
	Now        time.Time
	StatusCode int
	IsError    bool
	IsPanic    bool
	Path       string
}

// GetContextValues returns the values from the context.
func GetContextValues(ctx context.Context) (*ContextValues, error) {
	v, ok := ctx.Value(key).(*ContextValues)
	if !ok {
		return nil, errors.New("api value missing from context")
	}
	return v, nil
}

// GetTracerUID returns the trace id from the context.
func GetTracerUID(ctx context.Context) string {
	v, ok := ctx.Value(key).(*ContextValues)
	if !ok {
		return "deadbeef-dead-beef-aaaa-000000000000"
	}
	return v.TracerUID
}

// SetTracerUID makes sure that the trace id is accessible via this context
func SetTracerUID(ctx context.Context, traceid string) error {
	v, ok := ctx.Value(key).(*ContextValues)
	if !ok {
		return errors.New("api value missing from context")
	}
	v.TracerUID = traceid
	return nil
}

// SetStatusCode sets the status code back into the context.
func SetStatusCode(ctx context.Context, statusCode int) error {
	v, ok := ctx.Value(key).(*ContextValues)
	if !ok {
		return errors.New("api value missing from context")
	}
	v.StatusCode = statusCode
	return nil
}

// SetIsError sets the error code back into the context.
func SetIsError(ctx context.Context) error {
	v, ok := ctx.Value(key).(*ContextValues)
	if !ok {
		return errors.New("api value missing from context")
	}
	v.IsError = true
	return nil
}

// SetIsPanic sets a bool to see if this was a panic
func SetIsPanic(ctx context.Context) error {
	v, ok := ctx.Value(key).(*ContextValues)
	if !ok {
		return errors.New("api value missing from context")
	}
	v.IsPanic = true
	return nil
}

// SetPath removes right slash and things like :id, :uid, :alias, etc
// This is used for metrics later without the DYNAMIC part
func SetPath(ctx context.Context, path string) error {
	v, ok := ctx.Value(key).(*ContextValues)
	if !ok {
		return errors.New("api value missing from context")
	}
	v.Path = path
	return nil
}
