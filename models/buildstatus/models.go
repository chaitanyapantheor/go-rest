package buildstatus

import (
	"context"
	"fmt"
	"time"
	"unsafe"

	"github.com/chaitanyamaili/go_rest/models/buildstatus/db"
)

// BuildStatus represents the state of the current session
//
//swagger:model BuildStatus
type BuildStatus struct {
	// Primary Key
	// type: integer
	// example: 1
	ID string `json:"id"`
	// Sluggified label
	// example: status-label
	Alias string `json:"alias"`
	// Clean name
	// example: Status Name
	Name string `json:"name"`
	// Database created value
	// example: 2021-05-25T00:53:16.535668Z
	CreatedOn time.Time `json:"created_on"`
	// Database last updated value
	// example: 2021-05-25T00:53:16.535668Z
	UpdatedOn time.Time `json:"updated_on"`
	// Database soft delete value
	// example: 2021-05-25T00:53:16.535668Z
	// swagger:ignore
	DeletedOn *time.Time `json:"deleted_on,omitempty"`
}

// NewBuildStatus contains information needed to create a new NewStatus.
//
//swagger:model NewBuildStatus
type NewBuildStatus struct {
	// Sluggified label
	// in: string
	// required: true
	// example: new-status
	Alias string `json:"alias" validate:"required,slug"`
	// Clear readable name
	// in: string
	// required: true
	// example: New Status
	Name string `json:"name" validate:"required,notblank"`
}

// UpdateBuildStatus defines what information may be provided to
// modify an existing UpdateNewStatus. All fields are optional
// so clients can send just the fields they want changed. It uses pointer
// fields so we can differentiate between a field that was not provided
// and a field that was provided as explicitly blank. Normally we do not
// want to use pointers to basic types but we make exceptions around
// marshalling/unmarshalling.
//
//swagger:model UpdateBuildStatus
type UpdateBuildStatus struct {
	// Sluggified label
	// in: string
	// required: true
	// example: updated-status
	Alias *string `json:"alias" validate:"omitempty,required,slug"`
	// Clear readable name
	// in: string
	// example: Updated Status
	Name *string `json:"name" validate:"omitempty,required,notblank"`
}

func toStatus(dbRS db.BuildStatus) BuildStatus {
	p := (*BuildStatus)(unsafe.Pointer(&dbRS))
	return *p
}

func toStatusSlice(dbSRs []db.BuildStatus) []BuildStatus {
	rs := make([]BuildStatus, len(dbSRs))
	for i, dbSR := range dbSRs {
		rs[i] = toStatus(dbSR)
	}
	return rs
}

//------------------------------------------------------------------------
// Fake data generators
//------------------------------------------------------------------------

// GenerateFakeData return an array for NewBuildStatuses
func (ns NewBuildStatus) GenerateFakeData(num int) []NewBuildStatus {
	var data []NewBuildStatus
	for i := 0; i < num; i++ {
		data = append(data, ns.fakeData(i+1))
	}
	return data
}

// fakeData creates the fake record
func (ns NewBuildStatus) fakeData(counter int) NewBuildStatus {
	return NewBuildStatus{
		Alias: fmt.Sprintf("alias%d", counter),
		Name:  fmt.Sprintf("Label %d", counter),
	}
}

// Seed runs create methods from an array of new values
func (c Core) Seed(ctx context.Context, data []NewBuildStatus) error {
	now := time.Now().UTC()
	for _, ns := range data {
		if _, err := c.Create(ctx, ns, now); err != nil {
			return fmt.Errorf("error seeding build status: %w", err)
		}
	}

	return nil
}
