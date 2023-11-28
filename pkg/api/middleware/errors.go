package middleware

import (
	"context"
	"net/http"

	"github.com/chaitanyamaili/go_rest/pkg/api"
	"github.com/chaitanyamaili/go_rest/pkg/database"
	"github.com/chaitanyamaili/go_rest/pkg/validate"
	"go.uber.org/zap"
)

// Errors handles errors coming out of the call chain. It detects normal
// application errors which are used to respond to the client in a uniform way.
// Unexpected errors (status >= 500) are logged.
func Errors(log *zap.SugaredLogger) api.Middleware {

	// This is the actual middleware function to be executed.
	m := func(handler api.Handler) api.Handler {

		// Create the handler that will be attached in the middleware chain.
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

			v, err := api.GetContextValues(ctx)
			if err != nil {
				return api.NewShutdownError("api value missing from context")
			}

			// Run the next handler and catch any propagated error.
			err = handler(ctx, w, r)
			if err != nil {

				log.Errorw("CLIENT ERROR", "tracer_uid", v.TracerUID, "ERROR", err)

				// Build out the error response.
				var er api.ErrorResponse
				var status int

				// Set the error count for the request middleware
				_ = api.SetIsError(ctx)

				switch {
				case database.IsError(err):
					reqErr := database.GetError(err)
					er = api.ErrorResponse{
						Error: reqErr.Error(),
					}
					status = reqErr.Status

				case validate.IsFieldErrors(err):
					fieldErrors := validate.GetFieldErrors(err)
					errMsg := validate.GetCustomError(err)
					if errMsg == "" {
						errMsg = "data validation error"
					}
					er = api.ErrorResponse{
						Error:  errMsg,
						Fields: fieldErrors.Fields(),
					}
					status = http.StatusBadRequest

				case api.IsRequestError(err):
					reqErr := api.GetRequestError(err)
					er = api.ErrorResponse{
						Error: reqErr.Error(),
					}
					status = reqErr.Status

				default:
					status = http.StatusInternalServerError
					er = api.ErrorResponse{
						Error: err.Error(),
					}
				}

				// Respond with the error back to the client
				if err := api.Respond(ctx, w, er, status); err != nil {
					return err
				}

				// If we receive the shutdown err we need to return it
				// back to the base handler to shut down the service
				if ok := api.IsShutdown(err); ok {
					return err
				}
			}

			return nil
		}
		return h
	}
	return m
}
