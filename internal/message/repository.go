package message

import (
	"context"
	"errors"
	"fmt"

	dbx "github.com/go-ozzo/ozzo-dbx"
	"github.com/joinself/restful-client/internal/entity"
	"github.com/joinself/restful-client/pkg/dbcontext"
	"github.com/joinself/restful-client/pkg/log"
)

// Repository encapsulates the logic to access messages from the data source.
type Repository interface {
	// Get returns the message with the specified message ID.
	Get(ctx context.Context, connectionID int, id string) (entity.Message, error)
	// Count returns the number of messages.
	Count(ctx context.Context, connectionID, messagesSince int) (int, error)
	// Query returns the list of messages with the given offset and limit.
	Query(ctx context.Context, connection int, messagesSince int, offset, limit int) ([]entity.Message, error)
	// Create saves a new message in the storage.
	Create(ctx context.Context, message *entity.Message) error
	// Update updates the message with given ID in the storage.
	Update(ctx context.Context, message entity.Message) error
	// Delete removes the message with given ID from the storage.
	Delete(ctx context.Context, connectionID int, id string) error
}

// repository persists messages in database
type repository struct {
	db     *dbcontext.DB
	logger log.Logger
}

// NewRepository creates a new message repository
func NewRepository(db *dbcontext.DB, logger log.Logger) Repository {
	return repository{db, logger}
}

// Get reads the message with the specified ID from the database.
func (r repository) Get(ctx context.Context, connectionID int, jti string) (entity.Message, error) {
	var message entity.Message

	err := r.db.With(ctx).
		Select().
		From("message").
		Where(&dbx.HashExp{"jti": jti, "connection_id": connectionID}).
		One(&message)

	if &message == nil {
		return message, errors.New("message not found")
	}

	return message, err
}

// Create saves a new message record in the database.
// It returns the ID of the newly inserted message record.
func (r repository) Create(ctx context.Context, message *entity.Message) error {
	return r.db.With(ctx).Model(message).Insert()
}

// Update saves the changes to an message in the database.
func (r repository) Update(ctx context.Context, message entity.Message) error {
	return r.db.With(ctx).Model(&message).Update()
}

// Delete deletes an message with the specified ID from the database.
func (r repository) Delete(ctx context.Context, connectionID int, jti string) error {
	message, err := r.Get(ctx, connectionID, jti)
	if err != nil {
		return err
	}
	return r.db.With(ctx).Model(&message).Delete()
}

// Count returns the number of the message records in the database.
func (r repository) Count(ctx context.Context, connectionID, messagesSince int) (int, error) {
	var count int
	exp := dbx.NewExp("connection_id={:id}", dbx.Params{"id": connectionID})
	if messagesSince > 0 {
		exp = dbx.And(dbx.HashExp{"connection_id": connectionID}, dbx.NewExp(fmt.Sprintf("id>%d", messagesSince)))
	}

	err := r.db.With(ctx).
		Select("COUNT(*)").
		From("message").
		Where(exp).
		Row(&count)
	return count, err
}

// Query retrieves the message records with the specified offset and limit from the database.
func (r repository) Query(ctx context.Context, connection int, messagesSince int, offset, limit int) ([]entity.Message, error) {
	var messages []entity.Message
	exp := dbx.NewExp("connection_id={:id}", dbx.Params{"id": connection})
	if messagesSince > 0 {
		exp = dbx.And(dbx.HashExp{"connection_id": connection}, dbx.NewExp(fmt.Sprintf("id>%d", messagesSince)))
	}

	err := r.db.With(ctx).
		Select().
		OrderBy("id").
		Where(exp).
		Offset(int64(offset)).
		Limit(int64(limit)).
		OrderBy("created_at DESC").
		All(&messages)
	return messages, err
}
