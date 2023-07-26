package group

import (
	"context"
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/joinself/restful-client/internal/entity"
	"github.com/joinself/restful-client/internal/test"
	"github.com/joinself/restful-client/pkg/log"
	"github.com/stretchr/testify/assert"
)

func TestRepository(t *testing.T) {
	logger, _ := log.NewForTest()
	db := test.DB(t)
	test.ResetTables(t, db, "room_connection", "room")
	repo := NewRepository(db, logger)

	ctx := context.Background()

	// initial count
	count, err := repo.Count(ctx)
	assert.Nil(t, err)

	uid, _ := uuid.NewV4()
	gid := "gid" + uid.String()
	appID := "appID" + uid.String()

	// create
	room, err := repo.Create(ctx, entity.Room{
		GID:       gid,
		Appid:     appID,
		Name:      "group1",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	assert.Nil(t, err)
	count2, _ := repo.Count(ctx)
	assert.Equal(t, 1, count2-count)

	// get
	group, err := repo.Get(ctx, appID, room.ID)
	assert.Nil(t, err)
	assert.Equal(t, "group1", group.Name)

	// get unexisting
	_, err = repo.Get(ctx, appID, 0)
	assert.Equal(t, sql.ErrNoRows, err)

	// update
	group.Name = "group1 updated"
	err = repo.Update(ctx, group)
	assert.Nil(t, err)
	group, _ = repo.Get(ctx, appID, room.ID)
	assert.Equal(t, "group1 updated", group.Name)

	// query
	groups, err := repo.Query(ctx, appID, 0, count2)
	assert.Nil(t, err)
	assert.Equal(t, count2, len(groups))

	// membership
	members := repo.MemberIDs(ctx, room.ID)
	assert.Equal(t, 0, len(members))

	connectionID := rand.Intn(99999999)
	err = test.CreateConnection(ctx, db, connectionID)
	assert.Nil(t, err)
	err = repo.AddMember(ctx, entity.RoomConnection{
		RoomID:       room.ID,
		ConnectionID: connectionID,
	})
	assert.Nil(t, err)

	members = repo.MemberIDs(ctx, room.ID)
	assert.Equal(t, 1, len(members))

	// delete
	err = repo.Delete(ctx, group.ID)
	assert.Nil(t, err)
	_, err = repo.Get(ctx, appID, room.ID)
	assert.Equal(t, sql.ErrNoRows, err)

	// delete without relations
	room2, err := repo.Create(ctx, entity.Room{
		GID:       "random",
		Appid:     "app2",
		Name:      "group2",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	assert.Nil(t, err)
	count3, _ := repo.Count(ctx)
	assert.Equal(t, 1, count3)
	err = repo.Delete(ctx, room2.ID)
	count4, _ := repo.Count(ctx)
	assert.Equal(t, 0, count4)
}
