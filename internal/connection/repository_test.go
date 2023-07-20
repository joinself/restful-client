package connection

import (
	"context"
	"database/sql"
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
	test.ResetTables(t, db, "request", "fact", "message")
	test.ResetTables(t, db, "connection")
	repo := NewRepository(db, logger)

	ctx := context.Background()

	// initial count
	count, err := repo.Count(ctx)
	assert.Nil(t, err)

	id, _ := uuid.NewV4()
	connectionID := "selfID" + id.String()
	appID := "appID" + id.String()

	// create
	err = repo.Create(ctx, entity.Connection{
		// ID:        1,
		SelfID:    connectionID,
		AppID:     appID,
		Name:      "connection1",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	assert.Nil(t, err)
	count2, _ := repo.Count(ctx)
	assert.Equal(t, 1, count2-count)

	// get
	connection, err := repo.Get(ctx, appID, connectionID)
	assert.Nil(t, err)
	assert.Equal(t, "connection1", connection.Name)

	// get unexisting
	_, err = repo.Get(ctx, appID, "test0")
	assert.Equal(t, sql.ErrNoRows, err)

	// update
	connection.Name = "connection1 updated"
	err = repo.Update(ctx, connection)
	assert.Nil(t, err)
	connection, _ = repo.Get(ctx, appID, connectionID)
	assert.Equal(t, "connection1 updated", connection.Name)

	// query
	connections, err := repo.Query(ctx, appID, 0, count2)
	assert.Nil(t, err)
	assert.Equal(t, count2, len(connections))

	// delete
	err = repo.Delete(ctx, connection.ID)
	assert.Nil(t, err)
	_, err = repo.Get(ctx, appID, connectionID)
	assert.Equal(t, sql.ErrNoRows, err)
	err = repo.Delete(ctx, 1)
	assert.Equal(t, sql.ErrNoRows, err)
}
