package fact

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/qiangxue/go-rest-api/internal/entity"
	"github.com/qiangxue/go-rest-api/internal/test"
	"github.com/qiangxue/go-rest-api/pkg/log"
	"github.com/stretchr/testify/assert"
)

func TestRepository(t *testing.T) {
	logger, _ := log.NewForTest()
	db := test.DB(t)
	test.ResetTables(t, db, "fact")
	repo := NewRepository(db, logger)

	ctx := context.Background()
	connection := "connection"

	// initial count
	count, err := repo.Count(ctx)
	assert.Nil(t, err)

	// create a new connection
	err = test.CreateConnection(ctx, db, connection)
	assert.Nil(t, err)

	// create
	err = repo.Create(ctx, entity.Fact{
		ID:           "test1",
		ConnectionID: connection,
		Body:         "fact1",
		IAT:          time.Now(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	})
	assert.Nil(t, err)
	count2, _ := repo.Count(ctx)
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
		Body:         "fact1 updated",
		IAT:          time.Now(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	})
	assert.Nil(t, err)
	fact, _ = repo.Get(ctx, "test1")
	assert.Equal(t, "fact1 updated", fact.Body)

	// query
	facts, err := repo.Query(ctx, connection, 0, count2)
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