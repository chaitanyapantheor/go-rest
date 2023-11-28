package buildgrp

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/chaitanyamaili/go_rest/models/build"
	"github.com/chaitanyamaili/go_rest/pkg/api"
	"github.com/chaitanyamaili/go_rest/pkg/database"
)

// Handlers manages the set of repository endpoints.
type Handlers struct {
	Build build.Core
}

// Create adds a build test to the system.
//
// swagger:operation POST /build Build BuildCreate
//
// # Creates a new Build
//
// ---
// produces:
// - application/json
// responses:
//
//	  "201":
//		   "$ref": "#/responses/BuildRes"
//	  "400":
//		   "$ref": "#/responses/errorResponse400"
func (h Handlers) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	v, err := api.GetContextValues(ctx)
	if err != nil {
		return api.NewShutdownError("api value missing from context")
	}

	var nrs build.NewBuild
	if err := api.Decode(r, &nrs); err != nil {
		return fmt.Errorf("unable to decode payload: %w", err)
	}

	rs, err := h.Build.Create(ctx, nrs, v.Now)
	if err != nil {
		return err
	}

	return api.Respond(ctx, w, rs, http.StatusCreated)
}

// Query all the build records
//
// swagger:operation GET /build Build BuildQuery
//
// # This is the summary for listing builds
//
// ---
// produces:
// - application/json
// responses:
//
//	  "200":
//		   "$ref": "#/responses/BuildRes"
func (h Handlers) Query(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	pagi, err := database.PaginationParams(r)
	if err != nil {
		return err
	}

	rs, err := h.Build.Query(ctx, pagi)
	if err != nil {
		switch {
		case errors.Is(err, build.ErrNotFound):
			return api.NewRequestError(err, http.StatusNotFound)
		default:
			return fmt.Errorf("unable to query for Requesting Tech: %w", err)
		}
	}

	return api.Respond(ctx, w, rs, http.StatusOK)
}

// QueryByID from an individual id
//
// swagger:operation GET /build/{id} User BuildQueryById
//
// # Getting a single build by ID
//
// ---
// produces:
// - application/json
// responses:
//
//	  "200":
//		   "$ref": "#/responses/BuildRes"
//	  "400":
//		   "$ref": "#/responses/errorResponse400"
//	  "404":
//		   "$ref": "#/responses/errorResponse404"
func (h Handlers) QueryByID(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	id := api.Param(r, "id")

	rs, err := h.Build.QueryByID(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, build.ErrInvalidID):
			return api.NewRequestError(err, http.StatusBadRequest)
		case errors.Is(err, build.ErrNotFound):
			return api.NewRequestError(err, http.StatusNotFound)
		default:
			return fmt.Errorf("requesting tech id[%s]: %w", id, err)
		}
	}

	return api.Respond(ctx, w, []build.Build{rs}, http.StatusOK)
}
