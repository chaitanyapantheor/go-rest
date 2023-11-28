package middleware

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/chaitanyamaili/go_rest/pkg/api"
)

// Panics recovers from panics and converts the panic to an error so it is
// reported in Metrics and handled in Errors.
func Panics() api.Middleware {

	// This is the actual middleware function to be executed.
	m := func(handler api.Handler) api.Handler {

		// Create the handler that will be attached in the middleware chain.
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) (err error) {

			defer func(ctx context.Context) {
				if rec := recover(); rec != nil {

					trace := debug.Stack()
					err = fmt.Errorf("API PANIC [%v] TRACE:\n%s", rec, string(trace))

					_ = api.SetIsPanic(ctx)

					// metrics.AddPanics(ctx)
				}
			}(ctx)

			// Call the next handler and set its return value in the err variable.
			return handler(ctx, w, r)
		}
		return h
	}
	return m
}
