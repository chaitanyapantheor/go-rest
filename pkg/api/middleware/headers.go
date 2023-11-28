package middleware

import (
	"context"
	"net/http"

	"github.com/chaitanyamaili/go_rest/pkg/api"
	"github.com/chaitanyamaili/go_rest/pkg/validate"
)

// RequiredHeaders contains all required header fields
type RequiredHeaders struct {
	TracerUID string `json:"tracer_uid" validate:"header,uuid4"`
	UserUID   string `json:"user_uid" validate:"header,uuid4"`
	SiteUID   string `json:"site_uid" validate:"header,uuid4"`
	OrgUID    string `json:"org_uid" validate:"omitempty,uuid4"`
}

// Headers check to see if the minimum number of headers are set and valid
func Headers() api.Middleware {

	// This is the actual middleware function to be executed.
	m := func(handler api.Handler) api.Handler {

		// Create the handler that will be attached in the middleware chain.
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

			nrs := RequiredHeaders{
				TracerUID: r.Header.Get("tracer_uid"),
				OrgUID:    r.Header.Get("org_uid"),
				UserUID:   r.Header.Get("user_uid"),
				SiteUID:   r.Header.Get("site_uid"),
			}

			if err := validate.Check(nrs); err != nil {
				return err
			}

			// Call the next handler.
			return handler(ctx, w, r)
		}
		return h
	}
	return m
}
