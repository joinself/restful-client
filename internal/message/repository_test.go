package message

import (
	"context"
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/joinself/restful-client/internal/entity"
	"github.com/joinself/restful-client/internal/test"
	"github.com/joinself/restful-client/pkg/log"
	"github.com/stretchr/testify/assert"
)

func TestRepository(t *testing.T) {
	logger, _ := log.NewForTest()
	db := test.DB(t)
	test.ResetTables(t, db, "message")
	repo := NewRepository(db, logger)

	ctx := context.Background()
	// Generate random integer
	connection := rand.Intn(99999999)

	// initial count
	count, err := repo.Count(ctx, connection, 0)
	assert.Nil(t, err)

	// create a new connection
	err = test.CreateConnection(ctx, db, connection)
	assert.Nil(t, err)

	// create
	msg := entity.Message{
		ConnectionID: connection,
		Body:         "message1",
		JTI:          "jti",
		IAT:          time.Now(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err = repo.Create(ctx, &msg)
	assert.Nil(t, err)
	count2, _ := repo.Count(ctx, connection, 0)
	assert.Equal(t, 1, count2-count)

	// get
	message, err := repo.Get(ctx, connection, msg.JTI)
	assert.Nil(t, err)
	assert.Equal(t, "message1", message.Body)
	_, err = repo.Get(ctx, connection, "0")
	assert.Equal(t, sql.ErrNoRows, err)

	// update
	err = repo.Update(ctx, entity.Message{
		ID:           msg.ID,
		JTI:          msg.JTI,
		ConnectionID: connection,
		Body:         "message1 updated",
		IAT:          time.Now(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	})
	assert.Nil(t, err)
	message, _ = repo.Get(ctx, connection, msg.JTI)
	assert.Equal(t, "message1 updated", message.Body)

	// query
	messages, err := repo.Query(ctx, connection, 0, 0, count2)
	assert.Nil(t, err)
	assert.Equal(t, count2, len(messages))

	// delete
	err = repo.Delete(ctx, connection, msg.JTI)
	assert.Nil(t, err)
	_, err = repo.Get(ctx, connection, msg.JTI)
	assert.Equal(t, sql.ErrNoRows, err)
	err = repo.Delete(ctx, connection, msg.JTI)
	assert.Equal(t, sql.ErrNoRows, err)
}
