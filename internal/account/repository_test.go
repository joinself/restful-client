package account

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
	test.ResetTables(t, db, "account")
	repo := NewRepository(db, logger)

	ctx := context.Background()

	// initial count
	count, err := repo.Count(ctx)
	assert.Nil(t, err)

	id, _ := uuid.NewV4()
	username := "selfID" + id.String()
	password := "password" + id.String()

	// create
	err = repo.Create(ctx, entity.Account{
		UserName:  username,
		Password:  password,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	assert.Nil(t, err)
	count2, _ := repo.Count(ctx)
	assert.Equal(t, 1, count2-count)

	// get
	account, err := repo.Get(ctx, username, password)
	assert.Nil(t, err)

	// get unexisting
	_, err = repo.Get(ctx, "unexisting", "test0")
	assert.Equal(t, sql.ErrNoRows, err)

	// update
	updatedPassword := "updated password"
	account.Password = updatedPassword
	err = repo.Update(ctx, account)
	assert.Nil(t, err)
	_, err = repo.Get(ctx, username, password) // getting with old password with produce an error
	assert.Equal(t, "invalid password", err.Error())
	account, err = repo.Get(ctx, username, updatedPassword)
	assert.Nil(t, err)

	// delete
	err = repo.Delete(ctx, account.ID)
	assert.Nil(t, err)
	_, err = repo.Get(ctx, username, updatedPassword)
	assert.Equal(t, sql.ErrNoRows, err)
	err = repo.Delete(ctx, 1)
	assert.Equal(t, sql.ErrNoRows, err)
}
