// Package api for handling API requests
package api

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/dimfeld/httptreemux/v5"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// A Handler is a type that handles a http request within the framework
type Handler func(ctx context.Context, w http.ResponseWriter, r *http.Request) error

// API is the entrypoint into our application and what configures our context
// object for each of our http handlers. Feel free to add any configuration
// data/logic on this Api struct
type API struct {
	mux      *httptreemux.ContextMux
	otmux    http.Handler
	shutdown chan os.Signal
	mw       []Middleware
}

// NewAPI creates an Api value that handle a set of routes for the application
func NewAPI(shutdown chan os.Signal, mw ...Middleware) *API {

	// Create an OpenTelemetry HTTP Handler which wraps our router. This will start
	// the initial span and annotate it with information about the request/response.
	//
	// This is configured to use the W3C TraceContext standard to set the remote
	// parent if a client request includes the appropriate headers
	// https://w3c.github.io/trace-context/

	mux := httptreemux.NewContextMux()

	return &API{
		mux:      mux,
		otmux:    otelhttp.NewHandler(mux, "request"),
		shutdown: shutdown,
		mw:       mw,
	}
}

// SignalShutdown is used to gracefully shut down the app when an integrity
// issue is identified
func (a *API) SignalShutdown() {
	a.shutdown <- syscall.SIGTERM
}

// ServeHTTP implements the http.Handler interface. It's the entry point for
// all http traffic and allows the opentelemetry mux to run first to handle
// tracing. The opentelemetry mux then calls the application mux to handle
// application traffic.  See NewApi function above for implementation
func (a *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.otmux.ServeHTTP(w, r)
}

// Handle sets a handler function for a given HTTP method and path pair
// to the application server mux
func (a *API) Handle(method string, path string, handler Handler, mw ...Middleware) {

	// First wrap handler specific middleware around this handler
	handler = wrapMiddleware(mw, handler)

	// Add the api's general middleware to the handler chain.
	handler = wrapMiddleware(a.mw, handler)

	// Execute each specific request
	// The function to execute for each request.
	h := func(w http.ResponseWriter, r *http.Request) {

		// Pull the context from the request and
		// use it as a separate parameter.
		ctx := r.Context()

		// Capture the parent request span from the context.
		// Once we properly setup otel / tracing
		span := trace.SpanFromContext(ctx)

		// Set the context with the required values to
		// process the request.
		v := ContextValues{
			TracerUID: span.SpanContext().TraceID().String(),
			Now:       time.Now(),
		}
		ctx = context.WithValue(ctx, key, &v)

		// Register this path and tracer uid for metrics later on
		_ = SetPath(ctx, path)

		// Call the wrapped handler functions.
		if err := handler(ctx, w, r); err != nil {
			a.SignalShutdown()
			return
		}

	}

	a.mux.Handle(method, path, h)
}

// NotFound sets the mux handler to log 404s
func (a *API) NotFound(log *zap.SugaredLogger) {
	a.mux.NotFoundHandler = func(w http.ResponseWriter, r *http.Request) {
		statusCode := http.StatusNotFound
		url := r.URL.String()

		http.Error(w, fmt.Sprintf("path not found: %s", url), statusCode)
	}
}
