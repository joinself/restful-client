package connection

import (
	"context"
	"errors"

	dbx "github.com/go-ozzo/ozzo-dbx"
	"github.com/joinself/restful-client/internal/entity"
	"github.com/joinself/restful-client/pkg/dbcontext"
	"github.com/joinself/restful-client/pkg/log"
)

// Repository encapsulates the logic to access connections from the data source.
type Repository interface {
	// Get returns the connection with the specified connection ID.
	Get(ctx context.Context, appID, selfID string) (entity.Connection, error)
	// Count returns the number of connections.
	Count(ctx context.Context, appID string) (int, error)
	// Query returns the list of connections with the given offset and limit.
	Query(ctx context.Context, appid string, offset, limit int) ([]entity.Connection, error)
	// Create saves a new connection in the storage.
	Create(ctx context.Context, connection entity.Connection) error
	// Update updates the connection with given ID in the storage.
	Update(ctx context.Context, connection entity.Connection) error
	// Delete removes the connection with given ID from the storage.
	Delete(ctx context.Context, id int) error
}

// repository persists connections in database
type repository struct {
	db     *dbcontext.DB
	logger log.Logger
}

// NewRepository creates a new connection repository
func NewRepository(db *dbcontext.DB, logger log.Logger) Repository {
	return repository{db, logger}
}

// Get reads the connection with the specified ID from the database.
func (r repository) Get(ctx context.Context, appID, selfID string) (entity.Connection, error) {
	var connections []entity.Connection

	err := r.db.With(ctx).
		Select().
		OrderBy("id").
		Where(&dbx.HashExp{"selfid": selfID, "appid": appID}).
		All(&connections)

	if len(connections) == 0 {
		return entity.Connection{}, errors.New("sql: no rows in result set")
	}

	return connections[0], err
}

// Create saves a new connection record in the database.
// It returns the ID of the newly inserted connection record.
func (r repository) Create(ctx context.Context, connection entity.Connection) error {
	return r.db.With(ctx).Model(&connection).Insert()
}

// Update saves the changes to an connection in the database.
func (r repository) Update(ctx context.Context, connection entity.Connection) error {
	return r.db.With(ctx).Model(&connection).Update()
}

// Delete deletes an connection with the specified ID from the database.
func (r repository) Delete(ctx context.Context, id int) error {
	connection, err := r.getByID(ctx, id)
	if err != nil {
		return err
	}
	return r.db.With(ctx).Model(&connection).Delete()
}

// Count returns the number of the connection records in the database.
func (r repository) Count(ctx context.Context, appID string) (int, error) {
	var count int
	err := r.db.With(ctx).Select("COUNT(*)").
		From("connection").
		Where(&dbx.HashExp{"appid": appID}).
		Row(&count)
	return count, err
}

// Query retrieves the connection records with the specified offset and limit from the database.
func (r repository) Query(ctx context.Context, appid string, offset, limit int) ([]entity.Connection, error) {
	var connections []entity.Connection
	err := r.db.With(ctx).
		Select().
		Where(&dbx.HashExp{"appid": appid}).
		OrderBy("id").
		Offset(int64(offset)).
		OrderBy("created_at DESC").
		Limit(int64(limit)).
		All(&connections)
	return connections, err
}

func (r repository) getByID(ctx context.Context, id int) (entity.Connection, error) {
	var connection entity.Connection
	err := r.db.With(ctx).Select().Model(id, &connection)
	return connection, err
}
