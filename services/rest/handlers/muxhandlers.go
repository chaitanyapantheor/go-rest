// Package v1 handlers manages the different versions of the API.
package handlers

import (
	"context"
	"net/http"
	"os"
	"sync"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	"github.com/chaitanyamaili/go_rest/pkg/api"
	"github.com/chaitanyamaili/go_rest/pkg/api/middleware"
	v1 "github.com/chaitanyamaili/go_rest/services/rest/handlers/v1"
)

// Options represent optional parameters.
type Options struct {
	corsOrigin string
}

// APIMuxConfig contains all the mandatory systems required by handlers.
type APIMuxConfig struct {
	Shutdown chan os.Signal
	Log      *zap.SugaredLogger
	DB       *sqlx.DB
	RWMux    *sync.RWMutex
	Headers  bool
}

// APIMux constructs a http.Handler with all application routes defined.
func APIMux(cfg APIMuxConfig, options ...func(opts *Options)) http.Handler {
	var opts Options
	for _, option := range options {
		option(&opts)
	}

	// Construct the web.App which holds all routes as well as common Middleware.
	mw := make([]api.Middleware, 0, 6)
	mw = append(mw, middleware.Logger(cfg.Log))
	// mw = append(mw, middleware.Metrics())
	mw = append(mw, middleware.Errors(cfg.Log))
	if cfg.Headers {
		mw = append(mw, middleware.Headers())
	}
	mw = append(mw, middleware.Panics())
	a := api.NewAPI(
		cfg.Shutdown,
		mw...,
	)

	// Register the 404 path not found so that we can log it
	a.NotFound(cfg.Log)

	// Accept CORS 'OPTIONS' preflight requests if config has been provided.
	// Don't forget to apply the CORS middleware to the routes that need it.
	// Example Config: `conf:"default:https://MY_DOMAIN.COM"`
	if opts.corsOrigin != "" {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			return nil
		}
		a.Handle(http.MethodOptions, "", h, middleware.Cors(opts.corsOrigin))
	}

	// Load the v1 routes.
	v1.Routes(a, v1.Config{
		Log:   cfg.Log,
		DB:    cfg.DB,
		RWMux: cfg.RWMux,
	})

	return a
}
