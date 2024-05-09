package signature

import (
	"context"
	"errors"
	"fmt"

	dbx "github.com/go-ozzo/ozzo-dbx"
	"github.com/joinself/restful-client/internal/entity"
	"github.com/joinself/restful-client/pkg/dbcontext"
	"github.com/joinself/restful-client/pkg/log"
)

// Repository encapsulates the logic to access signatures from the data source.
type Repository interface {
	// Get returns the signature with the specified signature ID.
	Get(ctx context.Context, appID, selfID, id string) (entity.Signature, error)
	// Create saves a new signature in the storage.
	Create(ctx context.Context, signature *entity.Signature) error
	// Update updates the signature with given ID in the storage.
	Update(ctx context.Context, signature entity.Signature) error
	// Delete removes the signature with given ID from the storage.
	Delete(ctx context.Context, appID, selfID, id string) error
	// Count returns the number of the signatures records in the database.
	Count(ctx context.Context, aID, cID string, signaturesSince int) (int, error)
	// Query retrieves the signatures records with the specified offset and limit from the database.
	Query(ctx context.Context, aID, cID string, signaturesSince int, offset, limit int) ([]entity.Signature, error)
}

// repository persists signatures in database
type repository struct {
	db     *dbcontext.DB
	logger log.Logger
}

// NewRepository creates a new signature repository
func NewRepository(db *dbcontext.DB, logger log.Logger) Repository {
	return repository{db, logger}
}

// Get reads the signature with the specified ID from the database.
func (r repository) Get(ctx context.Context, appID, selfID, id string) (entity.Signature, error) {
	var signature entity.Signature

	err := r.db.With(ctx).
		Select().
		From("signature").
		Where(&dbx.HashExp{"id": id, "self_id": selfID, "app_id": appID}).
		One(&signature)

	if &signature == nil {
		return signature, errors.New("signature not found")
	}

	return signature, err
}

// Create saves a new signature record in the database.
// It returns the ID of the newly inserted signature record.
func (r repository) Create(ctx context.Context, signature *entity.Signature) error {
	return r.db.With(ctx).Model(signature).Insert()
}

// Update saves the changes to an signature in the database.
func (r repository) Update(ctx context.Context, signature entity.Signature) error {
	return r.db.With(ctx).Model(&signature).Update()
}

// Delete deletes an signature with the specified ID from the database.
func (r repository) Delete(ctx context.Context, appID, selfID, id string) error {
	signature, err := r.Get(ctx, appID, selfID, id)
	if err != nil {
		return err
	}
	return r.db.With(ctx).Model(&signature).Delete()
}

// Count returns the number of the signatures records in the database.
func (r repository) Count(ctx context.Context, aID, cID string, signaturesSince int) (int, error) {
	var count int
	exp := dbx.NewExp("self_id={:cid} AND app_id={:aid}", dbx.Params{"cid": cID, "aid": aID})
	if signaturesSince > 0 {
		exp = dbx.And(dbx.NewExp(fmt.Sprintf("id>%d", signaturesSince)))
	}

	err := r.db.With(ctx).
		Select("COUNT(*)").
		From("signature").
		Where(exp).
		Row(&count)
	return count, err
}

// Query retrieves the signatures records with the specified offset and limit from the database.
func (r repository) Query(ctx context.Context, aID, cID string, signaturesSince int, offset, limit int) ([]entity.Signature, error) {
	var signatures []entity.Signature
	exp := dbx.NewExp("self_id={:cid} AND app_id={:aid}", dbx.Params{"cid": cID, "aid": aID})
	if signaturesSince > 0 {
		exp = dbx.And(dbx.NewExp(fmt.Sprintf("id>%d", signaturesSince)))
	}

	err := r.db.With(ctx).
		Select().
		OrderBy("id").
		Where(exp).
		Offset(int64(offset)).
		Limit(int64(limit)).
		OrderBy("created_at DESC").
		All(&signatures)
	return signatures, err
}
