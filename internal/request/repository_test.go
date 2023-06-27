package request

import (
	"context"
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
	test.ResetTables(t, db, "request")
	repo := NewRepository(db, logger)

	ctx := context.Background()
	// Generate random integer
	connection := rand.Intn(99999999)

	// create a new connection
	err := test.CreateConnection(ctx, db, connection)
	assert.Nil(t, err)

	// create
	req := entity.Request{
		ConnectionID: connection,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err = repo.Create(ctx, req)
	assert.Nil(t, err)

	// get
	message, err := repo.Get(ctx, req.ID)
	assert.Nil(t, err)
	assert.Equal(t, connection, message.ConnectionID)

	// get unexisting
	req, err = repo.Get(ctx, "unexisting")
	assert.NotNil(t, err)
	assert.Equal(t, req.ID, "")
}
