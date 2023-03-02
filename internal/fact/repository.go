package fact

import (
	"context"

	dbx "github.com/go-ozzo/ozzo-dbx"
	"github.com/qiangxue/go-rest-api/internal/entity"
	"github.com/qiangxue/go-rest-api/pkg/dbcontext"
	"github.com/qiangxue/go-rest-api/pkg/log"
)

// Repository encapsulates the logic to access facts from the data source.
type Repository interface {
	// Get returns the fact with the specified fact ID.
	Get(ctx context.Context, id string) (entity.Fact, error)
	// Count returns the number of facts.
	Count(ctx context.Context) (int, error)
	// Query returns the list of facts with the given offset and limit.
	Query(ctx context.Context, connection string, offset, limit int) ([]entity.Fact, error)
	// Create saves a new fact in the storage.
	Create(ctx context.Context, fact entity.Fact) error
	// Update updates the fact with given ID in the storage.
	Update(ctx context.Context, fact entity.Fact) error
	// Delete removes the fact with given ID from the storage.
	Delete(ctx context.Context, id string) error
}

// repository persists facts in database
type repository struct {
	db     *dbcontext.DB
	logger log.Logger
}

// NewRepository creates a new fact repository
func NewRepository(db *dbcontext.DB, logger log.Logger) Repository {
	return repository{db, logger}
}

// Get reads the fact with the specified ID from the database.
func (r repository) Get(ctx context.Context, id string) (entity.Fact, error) {
	var fact entity.Fact
	err := r.db.With(ctx).Select().Model(id, &fact)
	return fact, err
}

// Create saves a new fact record in the database.
// It returns the ID of the newly inserted fact record.
func (r repository) Create(ctx context.Context, fact entity.Fact) error {
	return r.db.With(ctx).Model(&fact).Insert()
}

// Update saves the changes to an fact in the database.
func (r repository) Update(ctx context.Context, fact entity.Fact) error {
	return r.db.With(ctx).Model(&fact).Update()
}

// Delete deletes an fact with the specified ID from the database.
func (r repository) Delete(ctx context.Context, id string) error {
	fact, err := r.Get(ctx, id)
	if err != nil {
		return err
	}
	return r.db.With(ctx).Model(&fact).Delete()
}

// Count returns the number of the fact records in the database.
func (r repository) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.With(ctx).Select("COUNT(*)").From("fact").Row(&count)
	return count, err
}

// Query retrieves the fact records with the specified offset and limit from the database.
func (r repository) Query(ctx context.Context, connection string, offset, limit int) ([]entity.Fact, error) {
	var facts []entity.Fact
	err := r.db.With(ctx).
		Select().
		OrderBy("id").
		Where(&dbx.HashExp{"connection_id": connection}).
		Offset(int64(offset)).
		Limit(int64(limit)).
		All(&facts)
	return facts, err
}
