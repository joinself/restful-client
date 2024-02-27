package apikey

import (
	"context"
	"database/sql"
	"errors"
	"time"

	dbx "github.com/go-ozzo/ozzo-dbx"
	"github.com/joinself/restful-client/internal/entity"
	"github.com/joinself/restful-client/pkg/dbcontext"
	"github.com/joinself/restful-client/pkg/filter"
	"github.com/joinself/restful-client/pkg/log"
)

// Repository encapsulates the logic to access apikeys from the data source.
type Repository interface {
	// Get returns the apikey with the specified apikey ID.
	Get(ctx context.Context, appID string, id int) (entity.Apikey, error)
	// Count returns the number of apikeys.
	Count(ctx context.Context, appID string) (int, error)
	// Query returns the list of apikeys with the given offset and limit.
	Query(ctx context.Context, appid string, offset, limit int) ([]entity.Apikey, error)
	// Create saves a new apikey in the storage.
	Create(ctx context.Context, apikey *entity.Apikey) error
	// Update updates the apikey with given ID in the storage.
	Update(ctx context.Context, apikey entity.Apikey) error
	// Delete removes the apikey with given ID from the storage.
	Delete(ctx context.Context, id int) error
	PreloadDeleted(ctx context.Context) error
}

// repository persists apikeys in database
type repository struct {
	db      *dbcontext.DB
	checker *filter.Checker
	logger  log.Logger
}

// NewRepository creates a new apikey repository
func NewRepository(db *dbcontext.DB, checker *filter.Checker, logger log.Logger) Repository {
	return repository{db, checker, logger}
}

// Get reads the apikey with the specified ID from the database.
func (r repository) Get(ctx context.Context, appID string, id int) (entity.Apikey, error) {
	var apikeys []entity.Apikey

	err := r.db.With(ctx).
		Select().
		OrderBy("id").
		Where(&dbx.HashExp{"id": id, "appid": appID}).
		AndWhere(&dbx.HashExp{"deleted": false}).
		All(&apikeys)

	if len(apikeys) == 0 {
		return entity.Apikey{}, errors.New("sql: no rows in result set")
	}

	return apikeys[0], err
}

// Create saves a new apikey record in the database.
// It returns the ID of the newly inserted apikey record.
func (r repository) Create(ctx context.Context, apikey *entity.Apikey) error {
	return r.db.With(ctx).Model(apikey).Insert()
}

// Update saves the changes to an apikey in the database.
func (r repository) Update(ctx context.Context, apikey entity.Apikey) error {
	return r.db.With(ctx).Model(&apikey).Update()
}

// Delete deletes an apikey with the specified ID from the database.
func (r repository) Delete(ctx context.Context, id int) error {
	apikey, err := r.getByID(ctx, id)
	if err != nil {
		return err
	}

	if apikey.Deleted {
		return sql.ErrNoRows
	}

	apikey.Deleted = true
	apikey.DeletedAt = time.Now()

	r.checker.Add(apikey.ID)

	return r.db.With(ctx).Model(&apikey).Update()
}

// Count returns the number of the apikey records in the database.
func (r repository) Count(ctx context.Context, appID string) (int, error) {
	var count int
	err := r.db.With(ctx).Select("COUNT(*)").
		From("apikey").
		Where(&dbx.HashExp{"appid": appID}).
		AndWhere(&dbx.HashExp{"deleted": false}).
		Row(&count)
	return count, err
}

// Query retrieves the apikey records with the specified offset and limit from the database.
func (r repository) Query(ctx context.Context, appid string, offset, limit int) ([]entity.Apikey, error) {
	var apikeys []entity.Apikey
	err := r.db.With(ctx).
		Select().
		Where(&dbx.HashExp{"appid": appid}).
		AndWhere(&dbx.HashExp{"deleted": false}).
		OrderBy("id").
		Offset(int64(offset)).
		OrderBy("created_at DESC").
		Limit(int64(limit)).
		All(&apikeys)
	return apikeys, err
}

func (r repository) getByID(ctx context.Context, id int) (entity.Apikey, error) {
	var apikey entity.Apikey
	err := r.db.With(ctx).Select().Model(id, &apikey)
	return apikey, err
}

func (r repository) PreloadDeleted(ctx context.Context) error {
	var apikeys []entity.Apikey
	err := r.db.With(ctx).
		Select().
		Where(&dbx.HashExp{"deleted": true}).
		OrderBy("id").
		OrderBy("created_at DESC").
		All(&apikeys)
	if err != nil {
		return err
	}

	for i := 0; i < len(apikeys); i++ {
		r.checker.Add(apikeys[i].ID)
	}
	return nil
}
