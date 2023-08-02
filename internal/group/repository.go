package group

import (
	"context"
	"errors"

	dbx "github.com/go-ozzo/ozzo-dbx"
	"github.com/joinself/restful-client/internal/entity"
	"github.com/joinself/restful-client/pkg/dbcontext"
	"github.com/joinself/restful-client/pkg/log"
)

// Repository encapsulates the logic to access groups from the data source.
type Repository interface {
	// Get returns the group with the specified group ID.
	Get(ctx context.Context, appID string, id int) (entity.Room, error)
	// Get gets a group by GID
	GetByGID(ctx context.Context, appID, gid string) (entity.Room, error)
	// Count returns the number of groups.
	Count(ctx context.Context) (int, error)
	// Query returns the list of groups with the given offset and limit.
	Query(ctx context.Context, appid string, offset, limit int) ([]entity.Room, error)
	// Create saves a new group in the storage.
	Create(ctx context.Context, group entity.Room) (entity.Room, error)
	// Update updates the group with given ID in the storage.
	Update(ctx context.Context, group entity.Room) error
	// Delete removes the group with given ID from the storage.
	Delete(ctx context.Context, id int) error
	// MemberIDs returns a list of the connection ids for this group.
	MemberIDs(ctx context.Context, id int) []int
	// AddMember adds a member to the group
	AddMember(ctx context.Context, relation entity.RoomConnection) error
	// RemoveMember removes a member from the group.
	RemoveMember(ctx context.Context, relation entity.RoomConnection) error
	// RemoveMembers removes all members associated to a specific group
	RemoveMembers(ctx context.Context, id int) error
}

// repository persists groups in database
type repository struct {
	db     *dbcontext.DB
	logger log.Logger
}

// NewRepository creates a new group repository
func NewRepository(db *dbcontext.DB, logger log.Logger) Repository {
	return repository{db, logger}
}

// Get reads the group with the specified ID from the database.
func (r repository) Get(ctx context.Context, appID string, id int) (entity.Room, error) {
	var groups []entity.Room

	err := r.db.With(ctx).
		Select().
		OrderBy("id").
		Where(&dbx.HashExp{"id": id, "appid": appID}).
		All(&groups)

	if len(groups) == 0 {
		return entity.Room{}, errors.New("sql: no rows in result set")
	}

	return groups[0], err
}

// Create saves a new group record in the database.
// It returns the ID of the newly inserted group record.
func (r repository) Create(ctx context.Context, group entity.Room) (entity.Room, error) {
	err := r.db.With(ctx).Model(&group).Insert()
	return group, err
}

// Update saves the changes to an group in the database.
func (r repository) Update(ctx context.Context, group entity.Room) error {
	return r.db.With(ctx).Model(&group).Update()
}

// Delete deletes an group with the specified ID from the database.
func (r repository) Delete(ctx context.Context, id int) error {
	// Delete all related members
	_, err := r.db.With(ctx).Delete("room_connection", dbx.HashExp{"room_id": id}).Execute()
	if err != nil {
		return err
	}

	// Delete the group
	_, err = r.db.With(ctx).Delete("room", dbx.HashExp{"id": id}).Execute()
	return err
}

// Count returns the number of the group records in the database.
func (r repository) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.With(ctx).Select("COUNT(*)").From("room").Row(&count)
	return count, err
}

// Query retrieves the group records with the specified offset and limit from the database.
func (r repository) Query(ctx context.Context, appid string, offset, limit int) ([]entity.Room, error) {
	var groups []entity.Room
	err := r.db.With(ctx).
		Select().
		Where(&dbx.HashExp{"appid": appid}).
		OrderBy("id").
		Offset(int64(offset)).
		OrderBy("created_at DESC").
		Limit(int64(limit)).
		All(&groups)
	return groups, err
}

// MemberIDs returns a list of the connection ids for this group.
func (r repository) MemberIDs(ctx context.Context, id int) []int {
	var connections []entity.RoomConnection
	r.db.With(ctx).
		Select().
		Where(&dbx.HashExp{"room_id": id}).
		From("room_connection").
		All(&connections)

	var result []int
	for _, c := range connections {
		result = append(result, c.ConnectionID)
	}

	return result
}

// AddMember adds a member to the group
func (r repository) AddMember(ctx context.Context, relation entity.RoomConnection) error {
	return r.db.With(ctx).Model(&relation).Insert()
}

func (r repository) RemoveMember(ctx context.Context, relation entity.RoomConnection) error {
	return r.db.With(ctx).Model(&relation).Delete()
}

func (r repository) RemoveMembers(ctx context.Context, id int) error {
	_, err := r.db.
		With(ctx).
		Delete("room_connection", dbx.HashExp{"room_id": id}).
		Execute()
	return err
}

// Get gets a group by GID
func (r repository) GetByGID(ctx context.Context, appID, gid string) (entity.Room, error) {
	var groups []entity.Room

	err := r.db.With(ctx).
		Select().
		OrderBy("id").
		Where(&dbx.HashExp{"gid": gid, "appid": appID}).
		All(&groups)

	if len(groups) == 0 {
		return entity.Room{}, errors.New("sql: no rows in result set")
	}

	return groups[0], err
}
