package signature

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
	test.ResetTables(t, db, "signature")
	repo := NewRepository(db, logger)

	ctx := context.Background()
	appID, _ := uuid.NewV4()
	connection, _ := uuid.NewV4()
	signatureID, _ := uuid.NewV4()

	// check it does not exist
	_, err := repo.Get(ctx, appID.String(), connection.String(), signatureID.String())
	assert.Error(t, err)

	// create
	msg := entity.Signature{
		ID:          signatureID.String(),
		AppID:       appID.String(),
		SelfID:      connection.String(),
		Description: "hello",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	err = repo.Create(ctx, &msg)
	assert.Nil(t, err)

	// check it exists
	signature, err := repo.Get(ctx, appID.String(), connection.String(), signatureID.String())
	assert.NoError(t, err)
	assert.Equal(t, appID.String(), signature.AppID)
	assert.Equal(t, connection.String(), signature.SelfID)
	assert.Equal(t, signatureID.String(), signature.ID)
	assert.Equal(t, "", signature.Status)

	// update
	signature.Status = "updated_status"
	err = repo.Update(ctx, signature)
	assert.Nil(t, err)

	signature, err = repo.Get(ctx, appID.String(), connection.String(), signatureID.String())
	assert.NoError(t, err)
	assert.Equal(t, appID.String(), signature.AppID)
	assert.Equal(t, connection.String(), signature.SelfID)
	assert.Equal(t, signatureID.String(), signature.ID)
	assert.Equal(t, "updated_status", signature.Status)

	// delete
	err = repo.Delete(ctx, appID.String(), connection.String(), signatureID.String())
	assert.Nil(t, err)
	_, err = repo.Get(ctx, appID.String(), connection.String(), signatureID.String())
	assert.Equal(t, sql.ErrNoRows, err)
	err = repo.Delete(ctx, appID.String(), connection.String(), signatureID.String())
	assert.Equal(t, sql.ErrNoRows, err)
}
