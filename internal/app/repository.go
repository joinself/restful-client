package app

import (
	"context"

	"github.com/joinself/restful-client/internal/entity"
	"github.com/joinself/restful-client/pkg/dbcontext"
	"github.com/joinself/restful-client/pkg/log"
)

// Repository encapsulates the logic to access apps from the data source.
type Repository interface {
	// Get returns the app with the specified app ID.
	Get(ctx context.Context, appID string) (entity.App, error)
	// Count returns the number of apps.
	Count(ctx context.Context) (int, error)
	// Create saves a new app in the storage.
	Create(ctx context.Context, app entity.App) error
	// Update updates the app with given ID in the storage.
	Update(ctx context.Context, app entity.App) error
	// Delete removes the app with given ID from the storage.
	Delete(ctx context.Context, id string) error
	// List all the configured apps
	List(ctx context.Context) ([]entity.App, error)
}

// repository persists apps in database
type repository struct {
	db     *dbcontext.DB
	logger log.Logger
}

// NewRepository creates a new app repository
func NewRepository(db *dbcontext.DB, logger log.Logger) Repository {
	return repository{db, logger}
}

// Get reads the app with the specified ID from the database.
func (r repository) Get(ctx context.Context, id string) (entity.App, error) {
	return r.getByID(ctx, id)
}

// Create saves a new app record in the database.
// It returns the ID of the newly inserted app record.
func (r repository) Create(ctx context.Context, app entity.App) error {
	return r.db.With(ctx).Model(&app).Insert()
}

// Update saves the changes to an app in the database.
func (r repository) Update(ctx context.Context, app entity.App) error {
	return r.db.With(ctx).Model(&app).Update()
}

// Delete deletes an app with the specified ID from the database.
func (r repository) Delete(ctx context.Context, id string) error {
	app, err := r.getByID(ctx, id)
	if err != nil {
		return err
	}
	return r.db.With(ctx).Model(&app).Delete()
}

// Count returns the number of the app records in the database.
func (r repository) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.With(ctx).Select("COUNT(*)").From("app").Row(&count)
	return count, err
}

func (r repository) List(ctx context.Context) ([]entity.App, error) {
	var apps []entity.App
	err := r.db.With(ctx).
		Select().
		OrderBy("id").
		OrderBy("created_at DESC").
		All(&apps)
	return apps, err
}

func (r repository) getByID(ctx context.Context, id string) (entity.App, error) {
	var app entity.App
	err := r.db.With(ctx).Select().Model(id, &app)
	return app, err
}
