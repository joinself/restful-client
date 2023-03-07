package attestation

import (
	"context"

	dbx "github.com/go-ozzo/ozzo-dbx"
	"github.com/joinself/restful-client/internal/entity"
	"github.com/joinself/restful-client/pkg/dbcontext"
	"github.com/joinself/restful-client/pkg/log"
)

// Repository encapsulates the logic to access attestations from the data source.
type Repository interface {
	// Get returns the attestation with the specified attestation ID.
	Get(ctx context.Context, id string) (entity.Attestation, error)
	// Query returns the list of attestations with the given offset and limit.
	Query(ctx context.Context, factID string, offset, limit int) ([]entity.Attestation, error)
	// Create saves a new attestation in the storage.
	Create(ctx context.Context, attestation entity.Attestation) error
}

// repository persists attestations in database
type repository struct {
	db     *dbcontext.DB
	logger log.Logger
}

// NewRepository creates a new attestation repository
func NewRepository(db *dbcontext.DB, logger log.Logger) Repository {
	return repository{db, logger}
}

// Get reads the attestation with the specified ID from the database.
func (r repository) Get(ctx context.Context, id string) (entity.Attestation, error) {
	var attestation entity.Attestation
	err := r.db.With(ctx).Select().Model(id, &attestation)
	return attestation, err
}

// Create saves a new attestation record in the database.
// It returns the ID of the newly inserted attestation record.
func (r repository) Create(ctx context.Context, attestation entity.Attestation) error {
	return r.db.With(ctx).Model(&attestation).Insert()
}

// Query retrieves the attestation records with the specified offset and limit from the database.
func (r repository) Query(ctx context.Context, factID string, offset, limit int) ([]entity.Attestation, error) {
	var attestations []entity.Attestation
	err := r.db.With(ctx).
		Select().
		OrderBy("id").
		Where(&dbx.HashExp{"fact_id": factID}).
		Offset(int64(offset)).
		Limit(int64(limit)).
		All(&attestations)
	return attestations, err
}
