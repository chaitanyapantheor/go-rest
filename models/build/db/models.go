package db

import "time"

// Build represent the structure we need for moving data
// between the app and the database.
type Build struct {
	ID            string     `db:"id"`
	UUID          string     `db:"uuid"`
	Label         string     `db:"label"`
	CommitSha     string     `db:"commit_sha"`
	BuildStatusID string     `db:"build_status_id"`
	CreatedOn     time.Time  `db:"created_on"`
	UpdatedOn     time.Time  `db:"updated_on"`
	DeletedOn     *time.Time `db:"deleted_on"`
}
