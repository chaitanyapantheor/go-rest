package build

import (
	"context"
	"fmt"
	"time"
	"unsafe"

	"github.com/chaitanyamaili/go_rest/models/build/db"
)

// Build represents the state of the current session
//
//swagger:model Build
type Build struct {
	// Primary Key
	// type: integer
	// example: 1
	ID string `json:"id"`
	// Sluggified uuid
	// example: build-uuid
	UUID string `json:"uuid"`
	// Clean label
	// example: Build label
	Label string `json:"label"`
	// Clean commit sha
	// example: Build commut_sha
	CommitSha string `json:"commit_sha"`
	// Clean status id
	// example: Build status id
	BuildStatusID string `json:"build_status_id"`
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

// NewBuild contains information needed to create a new NewStatus.
//
//swagger:model NewBuild
type NewBuild struct {
	// Sluggified uuid
	// in: string
	// required: true
	// example: new-build
	UUID string `json:"uuid" validate:"required,slug"`
	// Sluggified label
	// in: string
	// required: true
	// example: new-build
	Label string `json:"label" validate:"required,slug"`
	// CommitSha
	// in: string
	// required: true
	// example: New Build
	CommitSha string `json:"commit_sha" validate:"required,notblank"`
	// StatusID
	// in: string
	// required: true
	// example: New Build
	BuildStatusID string `json:"build_status_id" validate:"required,notblank"`
}

// UpdateBuildStatus defines what information may be provided to
// modify an existing UpdateNewStatus. All fields are optional
// so clients can send just the fields they want changed. It uses pointer
// fields so we can differentiate between a field that was not provided
// and a field that was provided as explicitly blank. Normally we do not
// want to use pointers to basic types but we make exceptions around
// marshalling/unmarshalling.
//
//swagger:model UpdateBuild
type UpdateBuild struct {
	// Sluggified label
	// in: string
	// required: true
	// example: updated-status
	Label *string `json:"label" validate:"omitempty,required,slug"`
	// CommitSha
	// in: string
	// required: true
	// example: Update Build
	CommitSha *string `json:"commit_sha" validate:"omitempty,required,slug"`
	// StatusID
	// in: string
	// required: true
	// example: Update Build
	BuildStatusID *string `json:"build_status_id" validate:"omitempty,required,slug"`
}

func toStatus(dbRS db.Build) Build {
	p := (*Build)(unsafe.Pointer(&dbRS))
	return *p
}

func toStatusSlice(dbSRs []db.Build) []Build {
	rs := make([]Build, len(dbSRs))
	for i, dbSR := range dbSRs {
		rs[i] = toStatus(dbSR)
	}
	return rs
}

//------------------------------------------------------------------------
// Fake data generators
//------------------------------------------------------------------------

// GenerateFakeData return an array for NewBuildStatuses
func (ns NewBuild) GenerateFakeData(num int) []NewBuild {
	var data []NewBuild
	for i := 0; i < num; i++ {
		data = append(data, ns.fakeData(i+1))
	}
	return data
}

// fakeData creates the fake record
func (ns NewBuild) fakeData(counter int) NewBuild {
	return NewBuild{
		UUID:          fmt.Sprintf("uuid%d", counter),
		Label:         fmt.Sprintf("label %d", counter),
		CommitSha:     fmt.Sprintf("commit_sha %d", counter),
		BuildStatusID: fmt.Sprintf("build_status_id %d", counter),
	}
}

// Seed runs create methods from an array of new values
func (c Core) Seed(ctx context.Context, data []NewBuild) error {
	now := time.Now().UTC()
	for _, ns := range data {
		if _, err := c.Create(ctx, ns, now); err != nil {
			return fmt.Errorf("error seeding build status: %w", err)
		}
	}

	return nil
}
