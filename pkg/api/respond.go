package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

// SuccessResponse is the form used for API responses for success in the API.
type SuccessResponse struct {
	// Success
	//
	Success bool `json:"success"`
	// Timestamp
	//
	// example: 1234567
	Timestamp int64 `json:"timestamp"`
	// Data
	// in: body
	Data interface{} `json:"data,omitempty"`
	// Errors
	// in: body
	Errors interface{} `json:"errors,omitempty"`
}

// Respond returns json to client
func Respond(ctx context.Context, w http.ResponseWriter, data interface{}, statusCode int) error {

	ctx, span := otel.GetTracerProvider().Tracer("").Start(ctx, "pkg.api.respond")
	span.SetAttributes(attribute.Int("statusCode", statusCode))

	// Set the status code for the request logger middleware
	err := SetStatusCode(ctx, statusCode)
	if err != nil {
		return err
	}

	// If no data is provided, just return status code -- Always return something
	//if statusCode == http.StatusNoContent {
	//	w.WriteHeader(statusCode)
	//	return nil
	//}

	r := SuccessResponse{
		Success:   true,
		Timestamp: time.Now().UTC().Unix(),
		Data:      data,
	}
	// If it's an error, it does not need to re-marshal
	if reflect.TypeOf(data) == reflect.TypeOf(ErrorResponse{}) {
		r.Success = false
		r.Data = nil
		r.Errors = data
	}

	// Convert the response to json
	jd, err := json.Marshal(r)
	if err != nil {
		return err
	}

	// set the content type now that we know there was no marshal error
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	// Send the result back to the client
	if _, err := w.Write(jd); err != nil {
		return fmt.Errorf("write fail: %+v", jd)
	}

	return nil
}
