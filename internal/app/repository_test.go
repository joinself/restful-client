package app

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
	test.ResetTables(t, db, "fact", "request", "message")
	test.ResetTables(t, db, "app")
	repo := NewRepository(db, logger)

	ctx := context.Background()

	// initial count
	count, err := repo.Count(ctx)
	assert.Nil(t, err)

	id, _ := uuid.NewV4()
	appID := "appID" + id.String()
	secret := "secret"
	name := "name"

	// create
	err = repo.Create(ctx, entity.App{
		ID:           appID,
		DeviceSecret: secret,
		Name:         name,
		Env:          "test",
		Callback:     "callback",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	})
	assert.Nil(t, err)
	count2, _ := repo.Count(ctx)
	assert.Equal(t, 1, count2-count)

	// get
	app, err := repo.Get(ctx, appID)
	assert.Nil(t, err)
	assert.Equal(t, name, app.Name)
	assert.Equal(t, secret, app.DeviceSecret)

	// get unexisting
	_, err = repo.Get(ctx, "test0")
	assert.Equal(t, sql.ErrNoRows, err)

	// update
	app.Name = "app1 updated"
	err = repo.Update(ctx, app)
	assert.Nil(t, err)
	app, _ = repo.Get(ctx, appID)
	assert.Equal(t, "app1 updated", app.Name)

	// delete
	err = repo.Delete(ctx, app.ID)
	assert.Nil(t, err)
	_, err = repo.Get(ctx, appID)
	assert.Equal(t, sql.ErrNoRows, err)
	// delete unexisting
	err = repo.Delete(ctx, "unexisting")
	assert.Equal(t, sql.ErrNoRows, err)
}
