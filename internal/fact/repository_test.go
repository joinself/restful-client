package fact

import (
	"context"
	"database/sql"
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
	test.ResetTables(t, db, "attestation", "fact", "message", "connection")
	repo := NewRepository(db, logger)

	ctx := context.Background()
	connection := 1
	request := "request"

	// initial count
	facts, err := repo.Query(ctx, connection, "", "", 0, 1000)
	assert.Nil(t, err)
	count := len(facts)

	// create a new connection
	err = test.CreateConnection(ctx, db, connection)
	assert.Nil(t, err)

	err = test.CreateRequest(ctx, db, request, connection)
	assert.Nil(t, err)

	// create
	err = repo.Create(ctx, entity.Fact{
		ID:           "test1",
		ConnectionID: connection,
		RequestID:    &request,
		Body:         "fact1",
		IAT:          time.Now(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	})
	assert.Nil(t, err)
	facts, err = repo.Query(ctx, connection, "", "", 0, 1000)
	assert.Nil(t, err)
	count2 := len(facts)
	assert.Equal(t, 1, count2-count)

	// get
	fact, err := repo.Get(ctx, "test1")
	assert.Nil(t, err)
	assert.Equal(t, "fact1", fact.Body)
	_, err = repo.Get(ctx, "test0")
	assert.Equal(t, sql.ErrNoRows, err)

	// update
	err = repo.Update(ctx, entity.Fact{
		ID:           "test1",
		ConnectionID: connection,
		RequestID:    &request,
		Body:         "fact1 updated",
		IAT:          time.Now(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	})
	assert.Nil(t, err)
	fact, _ = repo.Get(ctx, "test1")
	assert.Equal(t, "fact1 updated", fact.Body)

	// query
	facts, err = repo.Query(ctx, connection, "", "", 0, count2)
	assert.Nil(t, err)

	assert.Equal(t, count2, len(facts))

	// delete
	err = repo.Delete(ctx, "test1")
	assert.Nil(t, err)
	_, err = repo.Get(ctx, "test1")
	assert.Equal(t, sql.ErrNoRows, err)
	err = repo.Delete(ctx, "test1")
	assert.Equal(t, sql.ErrNoRows, err)
}
