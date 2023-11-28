module github.com/chaitanyamaili/go_rest/models

go 1.19

replace github.com/chaitanyamaili/go_rest/pkg => ../pkg

require (
	github.com/chaitanyamaili/go_rest/pkg v0.0.0-00010101000000-000000000000
	github.com/jmoiron/sqlx v1.3.5
	go.uber.org/zap v1.26.0
)

require (
	cloud.google.com/go/compute v1.14.0 // indirect
	cloud.google.com/go/compute/metadata v0.2.3 // indirect
	github.com/dimfeld/httptreemux/v5 v5.5.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/go-logr/logr v1.3.0 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.46.1 // indirect
	go.opentelemetry.io/otel v1.21.0 // indirect
	go.opentelemetry.io/otel/metric v1.21.0 // indirect
	go.opentelemetry.io/otel/trace v1.21.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
)
