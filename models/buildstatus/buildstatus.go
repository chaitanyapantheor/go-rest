package buildstatus

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/chaitanyamaili/go_rest/models/buildstatus/db"
	"github.com/chaitanyamaili/go_rest/pkg/database"
	"github.com/chaitanyamaili/go_rest/pkg/validate"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound     = errors.New("build status not found")
	ErrInvalidID    = errors.New("ID is not in its proper form")
	ErrInvalidAlias = errors.New("alias is not in its proper form")
)

// Core manages the set of APIs for requesting source access
type Core struct {
	store db.Store
}

// NewCore constructs a core for requesting source api access.
func NewCore(log *zap.SugaredLogger, sqlxDB *sqlx.DB, rwmux *sync.RWMutex) Core {
	return Core{
		store: db.NewStore(log, sqlxDB, rwmux),
	}
}

// -----------------------------------------------------------------------
// CRUD Methods
// -----------------------------------------------------------------------

// Create inserts a new requesting source into the database
func (c Core) Create(ctx context.Context, rs NewBuildStatus, now time.Time) (BuildStatus, error) {
	if err := validate.Check(rs); err != nil {
		return BuildStatus{}, err
	}

	dbRS := db.BuildStatus{
		Alias:     strings.TrimSpace(rs.Alias),
		Name:      strings.TrimSpace(rs.Name),
		CreatedOn: now,
		UpdatedOn: now,
	}

	// This provides an example of how to execute a transaction if required.
	tran := func(tx sqlx.ExtContext) error {
		res, err := c.store.Tran(tx).Create(ctx, dbRS)
		if err != nil {
			return fmt.Errorf("create: %w", err)
		}
		dbRS.ID = fmt.Sprintf("%d", res.LastInsertID)
		return nil
	}

	if err := c.store.WithinTran(ctx, tran); err != nil {
		return BuildStatus{}, fmt.Errorf("tran: %w", err)
	}

	return toStatus(dbRS), nil
}

// Update replaces a requesting source document in the database.
func (c Core) Update(ctx context.Context, id string, urs UpdateBuildStatus, now time.Time) error {
	if err := validate.Check(urs); err != nil {
		return err
	}
	if err := validate.CheckID(id); err != nil {
		return ErrInvalidID
	}

	dbRS, err := c.store.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, database.ErrDBNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("updating status id[%s]: %w", id, err)
	}

	hasChanges := false
	if urs.Alias != nil {
		dbRS.Alias = strings.TrimSpace(*urs.Alias)
		hasChanges = true
	}
	if urs.Name != nil {
		dbRS.Name = strings.TrimSpace(*urs.Name)
		hasChanges = true
	}
	// No changes were made - don't touch the DB
	if !hasChanges {
		return nil
	}
	dbRS.UpdatedOn = now

	_, err = c.store.Update(ctx, dbRS)
	if err != nil {
		return fmt.Errorf("update id[%s]: %w", id, err)
	}

	return nil
}

// Delete removes a requesting source from the database.
func (c Core) Delete(ctx context.Context, id string, now time.Time) error {
	if err := validate.CheckID(id); err != nil {
		return ErrInvalidID
	}

	_, err := c.store.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, database.ErrDBNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("undeleting status id[%s]: %w", id, err)
	}

	_, err = c.store.Delete(ctx, id, now)
	if err != nil {
		return fmt.Errorf("delete id[%s]: %w", id, err)
	}

	return nil
}

// UnDelete removes a requesting source from the database.
func (c Core) UnDelete(ctx context.Context, id string) (BuildStatus, error) {
	if err := validate.CheckID(id); err != nil {
		return BuildStatus{}, ErrInvalidID
	}

	_, err := c.store.UnDelete(ctx, id)
	if err != nil {
		return BuildStatus{}, fmt.Errorf("undelete: %w", err)
	}

	dbRS, err := c.store.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, database.ErrDBNotFound) {
			return BuildStatus{}, ErrNotFound
		}
		return BuildStatus{}, fmt.Errorf("undeleting status id[%s]: %w", id, err)
	}

	return toStatus(dbRS), nil
}

// Query retrieves a list of existing records from the database
func (c Core) Query(ctx context.Context, pagi database.Pagination) ([]BuildStatus, error) {
	res, err := c.store.Query(ctx, pagi)
	if err != nil {
		return []BuildStatus{}, fmt.Errorf("query: %w", err)
	}

	return toStatusSlice(res), nil
}

// QueryByID retrieves a single records from the database by id
func (c Core) QueryByID(ctx context.Context, id string) (BuildStatus, error) {
	if err := validate.CheckID(id); err != nil {
		return BuildStatus{}, ErrInvalidID
	}

	res, err := c.store.QueryByID(ctx, id)
	if err != nil {
		if errors.Is(err, database.ErrDBNotFound) {
			return BuildStatus{}, ErrNotFound
		}
		return BuildStatus{}, fmt.Errorf("query: %w", err)
	}

	return toStatus(res), nil
}

// QueryByAlias retrieves a single records from the database by id
func (c Core) QueryByAlias(ctx context.Context, alias string) (BuildStatus, error) {
	if err := validate.CheckSlug(alias); err != nil {
		return BuildStatus{}, ErrInvalidAlias
	}

	res, err := c.store.QueryByAlias(ctx, alias)
	if err != nil {
		if errors.Is(err, database.ErrDBNotFound) {
			return BuildStatus{}, ErrNotFound
		}
		return BuildStatus{}, fmt.Errorf("query: %w", err)
	}

	return toStatus(res), nil
}
