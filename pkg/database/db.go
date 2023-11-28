package database

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"strings"
	"time"

	"cloud.google.com/go/compute/metadata"
	"github.com/chaitanyamaili/go_rest/pkg/api"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// Set of error variables for CRUD operations.
var (
	ErrDBNotFound        = errors.New("data not found")
	ErrDBDuplicatedEntry = errors.New("duplicated entry")
)

// Config is the required properties for the db
type Config struct {
	Type         string
	User         string
	Password     string
	Host         string
	Port         int
	Name         string
	MaxIdleConns int
	MaxOpenConns int
	DisableTLS   bool
}

// DBResults to store database operation results
type DBResults struct {
	LastInsertID int64
	AffectedRows int64
}

// -----------------------------------------------------------------------
// Connection Information
// -----------------------------------------------------------------------

// Open a connection to a db
func Open(cfg Config) (*sqlx.DB, error) {
	fmt.Println("database")
	cs := connectionString(cfg)

	db, err := sqlx.Open(cfg.Type, cs)
	if err != nil {
		return nil, err
	}

	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetMaxOpenConns(cfg.MaxOpenConns)

	return db, nil
}

func connectionString(cfg Config) string {

	switch strings.ToLower(cfg.Type) {
	case "mysql", "maria", "mariadb":
		return mysqlConnectionString(cfg)
	}

	return ""
}

func mysqlConnectionString(cfg Config) string {
	// For Local
	if !metadata.OnGCE() {
		q := make(url.Values)
		q.Set("parseTime", "true")

		if cfg.Port > 0 {
			cfg.Host = fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
		}

		//Weird mysql issue, host:port has to have parenthesis around it
		cfg.Host = "(" + cfg.Host + ")"

		u := url.URL{
			User:     url.UserPassword(cfg.User, cfg.Password),
			Host:     cfg.Host,
			Path:     cfg.Name,
			RawQuery: q.Encode(),
		}
		return strings.Trim(u.String(), "/")
	}
	// For CF in GCP
	u := fmt.Sprintf("%s:%s@unix(/%s)/%s?parseTime=true",
		cfg.User, cfg.Password, cfg.Host, cfg.Name)

	return strings.Trim(u, "/")
}

// -----------------------------------------------------------------------
// Debugging
// -----------------------------------------------------------------------

// StatusCheck returns nil if it can successfully talk to the database. It
// returns a non-nil error otherwise.
func StatusCheck(ctx context.Context, db *sqlx.DB) error {

	// First check we can ping the database.
	var pingError error
	for attempts := 1; ; attempts++ {
		pingError = db.Ping()
		if pingError == nil {
			break
		}
		time.Sleep(time.Duration(attempts) * 100 * time.Millisecond)
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}

	// Make sure we didn't timeout or be cancelled.
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// Run a simple query to determine connectivity. Running this query forces a
	// round trip through the database.
	const q = `SELECT true`
	var tmp bool
	return db.QueryRowContext(ctx, q).Scan(&tmp)
}

// -----------------------------------------------------------------------
// Transactions
// -----------------------------------------------------------------------

// Transactor interface needed to begin transaction.
type Transactor interface {
	Beginx() (*sqlx.Tx, error)
}

// WithinTran runs passed function and do commit/rollback at the end.
func WithinTran(ctx context.Context, log *zap.SugaredLogger, db Transactor, fn func(sqlx.ExtContext) error) error {
	traceID := api.GetTracerUID(ctx)

	// Begin the transaction.
	log.Infow("begin db transaction", "traceid", traceID)
	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("begin db transaction: %w", err)
	}

	// Mark to the defer function a rollback is required.
	mustRollback := true

	// Set up a defer function for rolling back the transaction. If
	// mustRollback is true it means the call to fn failed, and we
	// need to roll back the transaction.
	defer func() {
		if mustRollback {
			log.Infow("rollback db transaction", "traceid", traceID)
			if err := tx.Rollback(); err != nil {
				log.Errorw("unable to rollback db transaction", "traceid", traceID, "ERROR", err)
			}
		}
	}()

	// Execute the code inside the transaction. If the function
	// fails, return the error and the defer function will roll back.
	if err := fn(tx); err != nil {
		return fmt.Errorf("exec db transaction: %w", err)
	}

	// Disarm the deferred rollback.
	mustRollback = false

	// Commit the transaction.
	log.Infow("commit db transaction", "traceid", traceID)
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit db transaction: %w", err)
	}

	return nil
}

// -----------------------------------------------------------------------
// Query Helpers
// -----------------------------------------------------------------------

// NamedExecContext is a helper function to execute a CUD operation with
// logging and tracing.
func NamedExecContext(ctx context.Context, log *zap.SugaredLogger, db sqlx.ExtContext, query string, data interface{}) (DBResults, error) {
	q := queryString(query, data)
	traceID := api.GetTracerUID(ctx)
	log.Debugw("database.NamedExecContext", "traceid", traceID, "query", q)

	var dbres DBResults
	res, err := sqlx.NamedExecContext(ctx, db, query, data)
	if err != nil {
		return DBResults{}, err
	}

	lid, err := res.LastInsertId()
	if err != nil {
		return DBResults{}, err
	}
	dbres.LastInsertID = lid

	ra, err := res.RowsAffected()
	if err != nil {
		return DBResults{}, err
	}
	dbres.AffectedRows = ra

	if val, err := res.RowsAffected(); err != nil {
		dbres.AffectedRows = val
	}

	return dbres, err
}

// NamedQuerySlice is a helper function for executing queries that return a
// collection of data to be unmarshalled into a slice.
func NamedQuerySlice(ctx context.Context, log *zap.SugaredLogger, db sqlx.ExtContext, query string, data interface{}, dest interface{}) error {
	q := queryString(query, data)
	traceID := api.GetTracerUID(ctx)
	log.Debugw("database.NamedQuerySlice", "traceid", traceID, "query", q)
	val := reflect.ValueOf(dest)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Slice {
		return errors.New("must provide a pointer to a slice")
	}

	rows, err := sqlx.NamedQueryContext(ctx, db, query, data)
	if err != nil {
		return err
	}
	defer rows.Close() //nolint:all

	slice := val.Elem()
	for rows.Next() {
		v := reflect.New(slice.Type().Elem())
		if err := rows.StructScan(v.Interface()); err != nil && !strings.Contains(err.Error(), "unsupported Scan, storing driver.Value type <nil> into type *json.RawMessage") {
			return err
		}
		slice.Set(reflect.Append(slice, v.Elem()))
	}

	return nil
}

// NamedQueryStruct is a helper function for executing queries that return a
// single value to be unmarshalled into a struct type.
func NamedQueryStruct(ctx context.Context, log *zap.SugaredLogger, db sqlx.ExtContext, query string, data interface{}, dest interface{}) error {
	q := queryString(query, data)
	traceID := api.GetTracerUID(ctx)
	log.Debugw("database.NamedQuerySlice", "traceid", traceID, "query", q)

	rows, err := sqlx.NamedQueryContext(ctx, db, query, data)
	if err != nil {
		return err
	}
	defer rows.Close() //nolint:all

	if !rows.Next() {
		return ErrDBNotFound
	}

	if err := rows.StructScan(dest); err != nil && !strings.Contains(err.Error(), "unsupported Scan, storing driver.Value type <nil> into type *json.RawMessage") {
		return err
	}

	return nil
}

// queryString provides a pretty print version of the query and parameters.
func queryString(query string, args ...interface{}) string {
	if args[0] == nil {
		return query
	}
	query, params, err := sqlx.Named(query, args)
	if err != nil {
		return err.Error()
	}

	for _, param := range params {
		var value string
		switch v := param.(type) {
		case *string:
			value = fmt.Sprintf("%v", v)
			if v != nil {
				value = fmt.Sprintf(`'%s'`, *v)
			}
		case string, []byte:
			value = fmt.Sprintf(`'%s'`, v)
		case json.RawMessage:
			value = fmt.Sprintf(`'%s'`, string(v))
		default:
			value = fmt.Sprintf("%v", v)
		}
		query = strings.Replace(query, "?", value, 1)
	}

	singleSpacePattern := regexp.MustCompile(`\s\s+`)
	query = singleSpacePattern.ReplaceAllString(query, " ")

	return strings.Trim(query, " ")
}
