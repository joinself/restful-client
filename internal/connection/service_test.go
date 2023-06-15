package connection

import (
	"context"
	"testing"

	"github.com/joinself/restful-client/pkg/log"
	"github.com/joinself/restful-client/pkg/mock"
	"github.com/stretchr/testify/assert"
)

func TestCreateConnectionRequest_Validate(t *testing.T) {
	tests := []struct {
		name      string
		model     CreateConnectionRequest
		wantError bool
	}{
		{"success", CreateConnectionRequest{SelfID: "selfid"}, false},
		{"required", CreateConnectionRequest{}, true},
		{"too long", CreateConnectionRequest{SelfID: "1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.model.Validate()
			assert.Equal(t, tt.wantError, err != nil)
		})
	}
}

func TestUpdateConnectionRequest_Validate(t *testing.T) {
	tests := []struct {
		name      string
		model     UpdateConnectionRequest
		wantError bool
	}{
		{"success", UpdateConnectionRequest{Name: "test"}, false},
		{"required", UpdateConnectionRequest{Name: ""}, true},
		{"too long", UpdateConnectionRequest{Name: "1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.model.Validate()
			assert.Equal(t, tt.wantError, err != nil)
		})
	}
}

func Test_service_CRUD(t *testing.T) {
	logger, _ := log.NewForTest()
	factClients := map[string]FactService{}
	aclManagers := map[string]ACLManager{}

	s := NewService(&mock.ConnectionRepositoryMock{}, logger, factClients, aclManagers)

	ctx := context.Background()

	// initial count
	count, _ := s.Count(ctx)
	assert.Equal(t, 0, count)

	// successful creation
	id := "selfid"
	appid := "appid"
	connection, err := s.Create(ctx, appid, CreateConnectionRequest{SelfID: id})
	assert.Nil(t, err)
	assert.Equal(t, id, connection.SelfID)
	assert.Equal(t, "", connection.Name)
	assert.NotEmpty(t, connection.CreatedAt)
	assert.NotEmpty(t, connection.UpdatedAt)
	count, _ = s.Count(ctx)
	assert.Equal(t, 1, count)

	// validation error in creation
	_, err = s.Create(ctx, appid, CreateConnectionRequest{SelfID: ""})
	assert.NotNil(t, err)
	count, _ = s.Count(ctx)
	assert.Equal(t, 1, count)

	// unexpected error in creation
	_, err = s.Create(ctx, appid, CreateConnectionRequest{SelfID: "error"})
	assert.Equal(t, mock.ErrCRUD, err)
	count, _ = s.Count(ctx)
	assert.Equal(t, 1, count)

	_, _ = s.Create(ctx, appid, CreateConnectionRequest{SelfID: "test2"})

	// update
	connection, err = s.Update(ctx, appid, id, UpdateConnectionRequest{Name: "test updated"})
	assert.Nil(t, err)
	assert.Equal(t, "test updated", connection.Name)
	_, err = s.Update(ctx, appid, "none", UpdateConnectionRequest{Name: "test updated"})
	assert.NotNil(t, err)

	// validation error in update
	_, err = s.Update(ctx, appid, id, UpdateConnectionRequest{Name: ""})
	assert.NotNil(t, err)
	count, _ = s.Count(ctx)
	assert.Equal(t, 2, count)

	// unexpected error in update
	_, err = s.Update(ctx, appid, id, UpdateConnectionRequest{Name: "error"})
	assert.Equal(t, mock.ErrCRUD, err)
	count, _ = s.Count(ctx)
	assert.Equal(t, 2, count)

	// get
	_, err = s.Get(ctx, appid, "none")
	assert.NotNil(t, err)
	connection, err = s.Get(ctx, appid, id)
	assert.Nil(t, err)
	assert.Equal(t, "test updated", connection.Name)
	assert.Equal(t, id, connection.SelfID)

	// query
	connections, _ := s.Query(ctx, appid, 0, 0)
	assert.Equal(t, 2, len(connections))

	// delete
	_, err = s.Delete(ctx, appid, "none")
	assert.NotNil(t, err)
	connection, err = s.Delete(ctx, appid, id)
	assert.Nil(t, err)
	assert.Equal(t, id, connection.SelfID)
	count, _ = s.Count(ctx)
	assert.Equal(t, 1, count)
}
