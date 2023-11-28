package db

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/chaitanyamaili/go_rest/pkg/database"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// Store holds details for basic database needs
type Store struct {
	log          *zap.SugaredLogger
	tr           database.Transactor
	db           sqlx.ExtContext
	rwmux        *sync.RWMutex
	isWithinTran bool
}

// NewStore constructs a data for api access.
func NewStore(log *zap.SugaredLogger, db *sqlx.DB, rwmux *sync.RWMutex) Store {
	return Store{
		log:   log,
		tr:    db,
		db:    db,
		rwmux: rwmux,
	}
}

// WithinTran runs passes function and do commit/rollback at the end.
func (s Store) WithinTran(ctx context.Context, fn func(sqlx.ExtContext) error) error {
	if s.isWithinTran {
		return fn(s.db)
	}
	s.rwmux.Lock()
	err := database.WithinTran(ctx, s.log, s.tr, fn)
	s.rwmux.Unlock()

	return err
}

// Tran return new Store with transaction in it.
func (s Store) Tran(tx sqlx.ExtContext) Store {
	return Store{
		log:          s.log,
		tr:           s.tr,
		db:           tx,
		isWithinTran: true,
	}
}

// -----------------------------------------------------------------------
// Database Query Repository
// -----------------------------------------------------------------------

// Create inserts a new requesting into the database.
func (s Store) Create(ctx context.Context, rs Build) (database.DBResults, error) {
	const q = `
	INSERT INTO build
		(uuid, label, commit_sha, build_status_id, created_on, updated_on)
	VALUES
		(:uuid, :label, :commit_sha, :build_status_id, :created_on, :updated_on)`

	res, err := database.NamedExecContext(ctx, s.log, s.db, q, rs)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return database.DBResults{}, database.NewError(database.ErrDBDuplicatedEntry, http.StatusConflict)
		}

		return database.DBResults{}, fmt.Errorf("inserting requesting source: %w", err)
	}

	return res, nil
}

// Update replaces a requesting source record in the database.
func (s Store) Update(ctx context.Context, rs Build) (database.DBResults, error) {
	const q = `
	UPDATE
		status
	SET 
		alias = :alias,
		label = :label,
		description = :description,
		updated_on = :updated_on
	WHERE
		id = :id`

	res, err := database.NamedExecContext(ctx, s.log, s.db, q, rs)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return database.DBResults{}, database.NewError(database.ErrDBDuplicatedEntry, http.StatusConflict)
		}
		return database.DBResults{}, fmt.Errorf("updating Requesting Source ID[%s]: %w", rs.ID, err)
	}

	return res, nil
}

// Delete removes a requesting source from the database.
func (s Store) Delete(ctx context.Context, id string, now time.Time) (database.DBResults, error) {
	data := struct {
		ID        string    `db:"id"`
		DeletedOn time.Time `db:"deleted_on"`
	}{
		ID:        id,
		DeletedOn: now,
	}

	const q = `
	UPDATE
		build
	SET
		deleted_on = :deleted_on
	WHERE
		id = :id`

	res, err := database.NamedExecContext(ctx, s.log, s.db, q, data)
	if err != nil {
		return database.DBResults{}, fmt.Errorf("deleting requesting source id[%s]: %w", id, err)
	}

	return res, nil
}

// UnDelete removed the deleted_on timestamp for a requesting source from the database
func (s Store) UnDelete(ctx context.Context, id string) (database.DBResults, error) {
	data := struct {
		ID string `db:"id"`
	}{ID: id}

	const q = `
	UPDATE
		build
	SET
		deleted_on = null
	WHERE
		id = :id`

	res, err := database.NamedExecContext(ctx, s.log, s.db, q, data)
	if err != nil {
		return database.DBResults{}, fmt.Errorf("undeleting requesting source id[%s]: %w", id, err)
	}

	return res, nil
}

// Query retrieves a list of existing requesting source from the database.
func (s Store) Query(ctx context.Context, pagi database.Pagination) ([]Build, error) {
	q := database.PaginationQuery(pagi, `
	SELECT
		id,
		uuid,
	    label,
	    commit_sha,
	    build_status_id,
	    created_on,
	    updated_on,
	    deleted_on
	FROM
		build
	WHERE
		deleted_on is null
	ORDER BY
		:sort :direction,
		id :direction
	LIMIT
		:page,:per_page`)

	// Slice to hold results
	var res []Build
	if err := database.NamedQuerySlice(ctx, s.log, s.db, q, pagi, &res); err != nil {
		if database.IsError(err) && err.Error() == database.ErrDBNotFound.Error() {
			return []Build{}, database.ErrDBNotFound
		}
		return nil, fmt.Errorf("selecting requesting source: %w", err)
	}

	return res, nil
}

// QueryByID retrieves a list of existing requesting sources from the database.
func (s Store) QueryByID(ctx context.Context, id string) (Build, error) {
	data := struct {
		ID string `db:"id"`
	}{ID: id}
	const q = `
	SELECT
		id,
		uuid,
		label,
		commit_sha,
		build_status_id,
		created_on,
		updated_on,
		deleted_on
	FROM
		build
	WHERE
		id = :id
		and deleted_on is null`

	// Slice to hold results
	var res Build
	if err := database.NamedQueryStruct(ctx, s.log, s.db, q, data, &res); err != nil {
		// Empty Check (no results)
		if database.IsError(err) && err.Error() == database.ErrDBNotFound.Error() {
			return Build{}, database.ErrDBNotFound
		}
		return Build{}, fmt.Errorf("selecting status by ID[%q]: %w", id, err)
	}

	return res, nil
}

// QueryByAlias retrieves a list of existing requesting sources from the database.
func (s Store) QueryByAlias(ctx context.Context, alias string) (Build, error) {
	data := struct {
		Alias string `db:"alias"`
	}{Alias: alias}
	const q = `
	SELECT
		id,
		alias,
		label,
		description,
		created_on,
		updated_on,
		deleted_on
	FROM
		status
	WHERE
		alias = :alias
		and deleted_on is null`

	// Slice to hold results
	var res Build
	if err := database.NamedQueryStruct(ctx, s.log, s.db, q, data, &res); err != nil {
		if database.IsError(err) && err.Error() == database.ErrDBNotFound.Error() {
			return Build{}, database.ErrDBNotFound
		}
		return Build{}, fmt.Errorf("selecting status by alias[%q]: %w", alias, err)
	}

	return res, nil
}
