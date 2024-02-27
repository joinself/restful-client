package apikey

import (
	"context"
	"testing"

	"github.com/joinself/restful-client/internal/config"
	"github.com/joinself/restful-client/internal/entity"
	"github.com/joinself/restful-client/pkg/log"
	"github.com/joinself/restful-client/pkg/mock"
	"github.com/stretchr/testify/assert"
)

func TestCreateApiKeyRequest_Validate(t *testing.T) {
	tests := []struct {
		name      string
		model     CreateApiKeyRequest
		wantError bool
	}{
		{"success", CreateApiKeyRequest{Name: "name", Scope: "FULL"}, false},
		{"required", CreateApiKeyRequest{}, true},
		{"too long", CreateApiKeyRequest{Name: "1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.model.Validate()
			assert.Equal(t, tt.wantError, err != nil)
		})
	}
}

func TestUpdateApiKeyRequest_Validate(t *testing.T) {
	tests := []struct {
		name      string
		model     UpdateApiKeyRequest
		wantError bool
	}{
		{"success", UpdateApiKeyRequest{Name: "test"}, false},
		{"required", UpdateApiKeyRequest{Name: ""}, true},
		{"too long", UpdateApiKeyRequest{Name: "1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890"}, true},
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
	s := NewService(&config.Config{}, &mock.ApikeyRepositoryMock{}, logger)

	ctx := context.Background()
	appid := "appid"

	// initial count
	count, _ := s.Count(ctx, appid)
	assert.Equal(t, 0, count)

	// successful creation
	apikey, err := s.Create(ctx, appid, CreateApiKeyRequest{Name: "name"}, entity.User{
		ID:   "1",
		Name: "test",
	})
	assert.Nil(t, err)
	assert.NotEmpty(t, apikey.ID)
	assert.Equal(t, "name", apikey.Name)
	assert.NotEmpty(t, apikey.CreatedAt)
	assert.NotEmpty(t, apikey.UpdatedAt)
	count, _ = s.Count(ctx, appid)
	assert.Equal(t, 1, count)

	id := apikey.ID

	_, _ = s.Create(ctx, appid, CreateApiKeyRequest{Name: "test2"}, entity.User{
		ID:   "1",
		Name: "test",
	})

	// update
	apikey, err = s.Update(ctx, appid, id, UpdateApiKeyRequest{Name: "test updated"})
	assert.Nil(t, err)
	assert.Equal(t, "test updated", apikey.Name)
	_, err = s.Update(ctx, appid, 0, UpdateApiKeyRequest{Name: "test updated"})
	assert.NotNil(t, err)

	// get
	_, err = s.Get(ctx, appid, 0)
	assert.NotNil(t, err)
	apikey, err = s.Get(ctx, appid, id)

	assert.Nil(t, err)
	assert.Equal(t, "test updated", apikey.Name)
	assert.Equal(t, id, 99)

	// query
	apikeys, _ := s.Query(ctx, appid, 0, 0)
	assert.Equal(t, 2, len(apikeys))

	// delete
	_, err = s.Delete(ctx, appid, 0)
	assert.NotNil(t, err)
	apikey, err = s.Delete(ctx, appid, id)
	assert.Nil(t, err)
	count, _ = s.Count(ctx, appid)
	assert.Equal(t, 1, count)
}
