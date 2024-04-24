package voice

import (
	"context"
	"errors"
	"fmt"

	dbx "github.com/go-ozzo/ozzo-dbx"
	"github.com/joinself/restful-client/internal/entity"
	"github.com/joinself/restful-client/pkg/dbcontext"
	"github.com/joinself/restful-client/pkg/log"
)

// Repository encapsulates the logic to access calls from the data source.
type Repository interface {
	// Get returns the call with the specified call ID.
	Get(ctx context.Context, appID, selfID, id string) (entity.Call, error)
	// Create saves a new call in the storage.
	Create(ctx context.Context, call *entity.Call) error
	// Update updates the call with given ID in the storage.
	Update(ctx context.Context, call entity.Call) error
	// Delete removes the call with given ID from the storage.
	Delete(ctx context.Context, appID, selfID, id string) error
	// Count returns the number of the calls records in the database.
	Count(ctx context.Context, aID, cID string, callsSince int) (int, error)
	// Query retrieves the calls records with the specified offset and limit from the database.
	Query(ctx context.Context, aID, cID string, callsSince int, offset, limit int) ([]entity.Call, error)
}

// repository persists calls in database
type repository struct {
	db     *dbcontext.DB
	logger log.Logger
}

// NewRepository creates a new call repository
func NewRepository(db *dbcontext.DB, logger log.Logger) Repository {
	return repository{db, logger}
}

// Get reads the call with the specified ID from the database.
func (r repository) Get(ctx context.Context, appID, selfID, id string) (entity.Call, error) {
	var call entity.Call

	err := r.db.With(ctx).
		Select().
		From("call").
		Where(&dbx.HashExp{"call_id": id, "selfid": selfID, "appid": appID}).
		One(&call)

	if &call == nil {
		return call, errors.New("call not found")
	}

	return call, err
}

// Create saves a new call record in the database.
// It returns the ID of the newly inserted call record.
func (r repository) Create(ctx context.Context, call *entity.Call) error {
	return r.db.With(ctx).Model(call).Insert()
}

// Update saves the changes to an call in the database.
func (r repository) Update(ctx context.Context, call entity.Call) error {
	return r.db.With(ctx).Model(&call).Update()
}

// Delete deletes an call with the specified ID from the database.
func (r repository) Delete(ctx context.Context, appID, selfID, id string) error {
	call, err := r.Get(ctx, appID, selfID, id)
	if err != nil {
		return err
	}
	return r.db.With(ctx).Model(&call).Delete()
}

// Count returns the number of the calls records in the database.
func (r repository) Count(ctx context.Context, aID, cID string, callsSince int) (int, error) {
	var count int
	exp := dbx.NewExp("selfid={:cid} AND appid={:aid}", dbx.Params{"cid": cID, "aid": aID})
	if callsSince > 0 {
		exp = dbx.And(dbx.NewExp(fmt.Sprintf("id>%d", callsSince)))
	}

	err := r.db.With(ctx).
		Select("COUNT(*)").
		From("call").
		Where(exp).
		Row(&count)
	return count, err
}

// Query retrieves the calls records with the specified offset and limit from the database.
func (r repository) Query(ctx context.Context, aID, cID string, callsSince int, offset, limit int) ([]entity.Call, error) {
	var calls []entity.Call
	exp := dbx.NewExp("selfid={:cid} AND appid={:aid}", dbx.Params{"cid": cID, "aid": aID})
	if callsSince > 0 {
		exp = dbx.And(dbx.NewExp(fmt.Sprintf("id>%d", callsSince)))
	}

	err := r.db.With(ctx).
		Select().
		OrderBy("id").
		Where(exp).
		Offset(int64(offset)).
		Limit(int64(limit)).
		OrderBy("created_at DESC").
		All(&calls)
	return calls, err
}
