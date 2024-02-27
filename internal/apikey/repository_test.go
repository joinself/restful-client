package apikey

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/joinself/restful-client/internal/entity"
	"github.com/joinself/restful-client/internal/test"
	"github.com/joinself/restful-client/pkg/filter"
	"github.com/joinself/restful-client/pkg/log"
	"github.com/stretchr/testify/assert"
)

func TestRepository(t *testing.T) {
	tokenChecker := filter.NewChecker()
	logger, _ := log.NewForTest()
	db := test.DB(t)
	test.ResetTables(t, db, "fact", "request", "message")
	test.ResetTables(t, db, "apikey")
	repo := NewRepository(db, tokenChecker, logger)

	ctx := context.Background()

	id, _ := uuid.NewV4()
	appID := "appID" + id.String()

	// initial count
	count, err := repo.Count(ctx, appID)
	assert.Nil(t, err)

	// create
	ak := entity.Apikey{
		// ID:        1,
		AppID:     appID,
		Token:     "sk...1234",
		Name:      "apikey1",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = repo.Create(ctx, &ak)
	assert.Nil(t, err)
	count2, _ := repo.Count(ctx, appID)
	assert.Equal(t, 1, count2-count)
	apikeyID := ak.ID

	// get
	apikey, err := repo.Get(ctx, appID, apikeyID)
	assert.Nil(t, err)

	// get unexisting
	_, err = repo.Get(ctx, appID, 2)
	assert.Equal(t, sql.ErrNoRows, err)

	// update
	apikey.Name = "apikey1 updated"
	err = repo.Update(ctx, apikey)
	assert.Nil(t, err)
	apikey, _ = repo.Get(ctx, appID, apikeyID)
	assert.Equal(t, "apikey1 updated", apikey.Name)

	// query
	apikeys, err := repo.Query(ctx, appID, 0, count2)
	assert.Nil(t, err)
	assert.Equal(t, count2, len(apikeys))

	// delete
	err = repo.Delete(ctx, apikey.ID)
	assert.Nil(t, err)
	_, err = repo.Get(ctx, appID, apikeyID)
	assert.Equal(t, sql.ErrNoRows, err)
	err = repo.Delete(ctx, 1)
	assert.Equal(t, sql.ErrNoRows, err)
}
