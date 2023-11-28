package v1

import (
	"net/http"
	"sync"

	"github.com/chaitanyamaili/go_rest/models/build"
	"github.com/chaitanyamaili/go_rest/pkg/api"
	"github.com/chaitanyamaili/go_rest/services/rest/handlers/v1/buildgrp"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Log   *zap.SugaredLogger
	DB    *sqlx.DB
	RWMux *sync.RWMutex
}

// Routes binds all the version 1 routes.
func Routes(api *api.API, cfg Config) {
	// -------------------------------------------------------------------
	// Build
	// -------------------------------------------------------------------
	bd := buildgrp.Handlers{
		Build: build.NewCore(cfg.Log, cfg.DB, cfg.RWMux),
	}
	api.Handle(http.MethodPost, "/v1/build", bd.Create)
	api.Handle(http.MethodGet, "/v1/build", bd.Query)
	api.Handle(http.MethodGet, "/v1/build/:id", bd.QueryByID)
}
