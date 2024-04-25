package voice

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
	test.ResetTables(t, db, "call")
	repo := NewRepository(db, logger)

	ctx := context.Background()
	appID, _ := uuid.NewV4()
	connection, _ := uuid.NewV4()
	callID, _ := uuid.NewV4()

	// check it does not exist
	_, err := repo.Get(ctx, appID.String(), connection.String(), callID.String())
	assert.Error(t, err)

	// create
	msg := entity.Call{
		AppID:     appID.String(),
		SelfID:    connection.String(),
		CallID:    callID.String(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = repo.Create(ctx, &msg)
	assert.Nil(t, err)

	// check it exists
	call, err := repo.Get(ctx, appID.String(), connection.String(), callID.String())
	assert.NoError(t, err)
	assert.Equal(t, appID.String(), call.AppID)
	assert.Equal(t, connection.String(), call.SelfID)
	assert.Equal(t, callID.String(), call.CallID)
	assert.Equal(t, "", call.Status)
	assert.Equal(t, "", call.PeerInfo)

	// update
	call.Status = "updated_status"
	call.PeerInfo = "updated_peer_info"
	err = repo.Update(ctx, call)
	assert.Nil(t, err)

	call, err = repo.Get(ctx, appID.String(), connection.String(), callID.String())
	assert.NoError(t, err)
	assert.Equal(t, appID.String(), call.AppID)
	assert.Equal(t, connection.String(), call.SelfID)
	assert.Equal(t, callID.String(), call.CallID)
	assert.Equal(t, "updated_status", call.Status)
	assert.Equal(t, "updated_peer_info", call.PeerInfo)

	// delete
	err = repo.Delete(ctx, appID.String(), connection.String(), callID.String())
	assert.Nil(t, err)
	_, err = repo.Get(ctx, appID.String(), connection.String(), callID.String())
	assert.Equal(t, sql.ErrNoRows, err)
	err = repo.Delete(ctx, appID.String(), connection.String(), callID.String())
	assert.Equal(t, sql.ErrNoRows, err)
}
