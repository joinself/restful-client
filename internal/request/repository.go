package request

import (
	"context"
	"time"

	"github.com/joinself/restful-client/internal/entity"
	"github.com/joinself/restful-client/pkg/dbcontext"
	"github.com/joinself/restful-client/pkg/log"
)

// Repository encapsulates the logic to access requests from the data source.
type Repository interface {
	// Get returns the request with the specified request ID.
	Get(ctx context.Context, id string) (entity.Request, error)
	// Create saves a new request in the storage.
	Create(ctx context.Context, request entity.Request) error
	// Update updates the request with given ID in the storage.
	Update(ctx context.Context, request entity.Request) error
	// Delete deletes an request with the specified ID from the database.
	Delete(ctx context.Context, id string) error
	// SetStatus updates the status of the given request.
	SetStatus(ctx context.Context, id string, status string) error
}

// repository persists requests in database
type repository struct {
	db     *dbcontext.DB
	logger log.Logger
}

// NewRepository creates a new request repository
func NewRepository(db *dbcontext.DB, logger log.Logger) Repository {
	return repository{db, logger}
}

// Get reads the request with the specified ID from the database.
func (r repository) Get(ctx context.Context, id string) (entity.Request, error) {
	var request entity.Request
	err := r.db.With(ctx).Select().Model(id, &request)
	return request, err
}

// Create saves a new request record in the database.
// It returns the ID of the newly inserted request record.
func (r repository) Create(ctx context.Context, request entity.Request) error {
	return r.db.With(ctx).Model(&request).Insert()
}

// Update saves the changes to an request in the database.
func (r repository) Update(ctx context.Context, request entity.Request) error {
	return r.db.With(ctx).Model(&request).Update()
}

// Delete deletes an request with the specified ID from the database.
func (r repository) Delete(ctx context.Context, id string) error {
	request, err := r.Get(ctx, id)
	if err != nil {
		return err
	}
	return r.db.With(ctx).Model(&request).Delete()
}

func (r repository) SetStatus(ctx context.Context, id string, status string) error {
	request, err := r.Get(ctx, id)
	if err != nil {
		return err
	}
	request.Status = status
	request.UpdatedAt = time.Now()
	return r.Update(ctx, request)
}
