package connection

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
	test.ResetTables(t, db, "connection")
	repo := NewRepository(db, logger)

	ctx := context.Background()

	// initial count
	count, err := repo.Count(ctx)
	assert.Nil(t, err)

	// create
	err = repo.Create(ctx, entity.Connection{
		ID:        "test1",
		Name:      "connection1",
		Selfid:    "0111111111",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	assert.Nil(t, err)
	count2, _ := repo.Count(ctx)
	println("..........")
	println(count)
	println(count2)
	println("..........")
	assert.Equal(t, 1, count2-count)

	// get
	connection, err := repo.Get(ctx, "test1")
	assert.Nil(t, err)
	assert.Equal(t, "connection1", connection.Name)
	_, err = repo.Get(ctx, "test0")
	assert.Equal(t, sql.ErrNoRows, err)

	// update
	err = repo.Update(ctx, entity.Connection{
		ID:        "test1",
		Name:      "connection1 updated",
		Selfid:    "1111111111",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	assert.Nil(t, err)
	connection, _ = repo.Get(ctx, "test1")
	assert.Equal(t, "connection1 updated", connection.Name)

	// query
	connections, err := repo.Query(ctx, 0, count2)
	assert.Nil(t, err)
	assert.Equal(t, count2, len(connections))

	// delete
	err = repo.Delete(ctx, "test1")
	assert.Nil(t, err)
	_, err = repo.Get(ctx, "test1")
	assert.Equal(t, sql.ErrNoRows, err)
	err = repo.Delete(ctx, "test1")
	assert.Equal(t, sql.ErrNoRows, err)
}
