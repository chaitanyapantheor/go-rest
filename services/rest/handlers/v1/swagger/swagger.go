// Package swagger for swagger document handler function
package swagger

import (
	"bytes"
	"context"
	"html/template"
	"net/http"

	mt "cloud.google.com/go/compute/metadata"
	"go.uber.org/zap"
)

// Handlers manages the set of check endpoints
type Handlers struct {
	Log *zap.SugaredLogger
}

// RedocOpts configures the Redoc middlewares
type RedocOpts struct {
	// BasePath for the UI path, defaults to: /
	BasePath string
	// Path combines with BasePath for the full UI path, defaults to: docs
	Path string
	// SpecURL the url to find the spec for
	SpecURL string
	// RedocURL for the js that generates the redoc site, defaults to: https://cdn.jsdelivr.net/npm/redoc/bundles/redoc.standalone.js
	RedocURL string
	// Title for the documentation site, default to: API documentation
	Title string
}

// EnsureDefaults in case some options are missing
func (r *RedocOpts) EnsureDefaults(req *http.Request) {
	if r.BasePath == "" {
		r.BasePath = "/"
	}
	if r.Path == "" {
		r.Path = "docs"
	}
	if r.SpecURL == "" {
		if mt.OnGCE() {
			r.SpecURL = "https://" + req.Host + "/docs/swagger.json"
		} else {
			r.SpecURL = "/docs/swagger.json"
		}
	}
	if r.RedocURL == "" {
		r.RedocURL = redocLatest
	}
	if r.Title == "" {
		r.Title = "API documentation"
	}
}

// Swagger returns a swagger json file
func (h Handlers) Swagger(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var ro RedocOpts
	ro.EnsureDefaults(r)
	tmpl := template.Must(template.New("redoc").Parse(redocTemplate))

	buf := bytes.NewBuffer(nil)
	_ = tmpl.Execute(buf, ro)
	b := buf.Bytes()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(b)

	return nil
}

func (h Handlers) ServeSwagger(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	http.ServeFile(w, r, "./swagger.json")

	return nil
}

const (
	redocLatest   = "https://cdn.jsdelivr.net/npm/redoc/bundles/redoc.standalone.js"
	redocTemplate = `<!DOCTYPE html>
<html>
  <head>
    <title>{{ .Title }}</title>
		<!-- needed for adaptive design -->
		<meta charset="utf-8"/>
		<meta name="viewport" content="width=device-width, initial-scale=1">
		<link href="https://fonts.googleapis.com/css?family=Montserrat:300,400,700|Roboto:300,400,700" rel="stylesheet">
    <!--
    ReDoc doesn't change outer page styles
    -->
    <style>
      body {
        margin: 0;
        padding: 0;
      }
    </style>
  </head>
  <body>
    <redoc spec-url='{{ .SpecURL }}'></redoc>
    <script src="{{ .RedocURL }}"> </script>
  </body>
</html>
`
)
