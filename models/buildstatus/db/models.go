package db

import "time"

// BuildStatus represent the structure we need for moving data
// between the app and the database.
type BuildStatus struct {
	ID        string     `db:"id"`
	Alias     string     `db:"alias"`
	Name      string     `db:"name"`
	CreatedOn time.Time  `db:"created_on"`
	UpdatedOn time.Time  `db:"updated_on"`
	DeletedOn *time.Time `db:"deleted_on"`
}
