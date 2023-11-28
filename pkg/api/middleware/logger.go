package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/chaitanyamaili/go_rest/pkg/api"
	"go.uber.org/zap"
)

// Logger writes some information about the request to the logs in the
// format: TraceID : (200) GET /foo -> IP ADDR (latency)
func Logger(log *zap.SugaredLogger) api.Middleware {
	// This is the actual middleware function to be executed.
	m := func(handler api.Handler) api.Handler {

		// Create the handler that will be attached in the middleware chain.
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

			v, err := api.GetContextValues(ctx)
			if err != nil {
				return api.NewShutdownError("api value missing from context")
			}

			start := time.Now()
			ou := []string{""}
			cn := ""
			if r.TLS != nil && len(r.TLS.VerifiedChains) > 0 && len(r.TLS.VerifiedChains[0]) > 0 {
				ou = r.TLS.VerifiedChains[0][0].Subject.OrganizationalUnit
				cn = r.TLS.VerifiedChains[0][0].Subject.CommonName
			}

			lw := log.With("component", "middleware:logger",
				"tracer_uid", v.TracerUID,
				"method", r.Method,
				"uri", r.RequestURI,
				"remote_addr", r.RemoteAddr,
				"user_agent", r.UserAgent(),
				"client_ou", ou,
				"client_cn", cn,
			)
			lw.Infow("request started")

			// Call the next handler
			err = handler(ctx, w, r)

			latency := time.Since(start)
			s := float64(latency.Microseconds()) / float64(1000000)
			lw.Infow("request completed",
				"duration_s", s,
				"response", v.StatusCode,
			)

			return err
		}
		return h
	}
	return m
}
