package app

import (
	"context"
	"testing"

	"github.com/joinself/restful-client/pkg/log"
	"github.com/joinself/restful-client/pkg/mock"
	"github.com/stretchr/testify/assert"
)

/*
	func TestCreateAppRequest_Validate(t *testing.T) {
		tests := []struct {
			name      string
			model     CreateAppRequest
			wantError bool
		}{
			{"success", CreateAppRequest{ID: "selfid"}, false},
			{"required", CreateAppRequest{}, true},
			{"too long", CreateAppRequest{ID: "1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890"}, true},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := tt.model.Validate()
				assert.Equal(t, tt.wantError, err != nil)
			})
		}
	}
*/
func Test_service_CRUD(t *testing.T) {
	logger, _ := log.NewForTest()
	runner := mock.NewRunnerMock()

	s := NewService(&mock.AppRepositoryMock{}, runner, logger)

	ctx := context.Background()

	// successful creation
	id := "appID"
	secret := "secret"
	name := "name"
	env := "env"
	callback := "callback"
	app, err := s.Create(ctx, CreateAppRequest{
		ID:       id,
		Secret:   secret,
		Name:     name,
		Env:      env,
		Callback: callback,
	})
	assert.Nil(t, err)
	assert.Equal(t, id, app.ID)
	assert.Equal(t, name, app.Name)
	assert.NotEmpty(t, app.CreatedAt)
	assert.NotEmpty(t, app.UpdatedAt)

	// validation error in creation
	_, err = s.Create(ctx, CreateAppRequest{
		ID:       "",
		Secret:   secret,
		Name:     name,
		Env:      env,
		Callback: callback,
	})
	assert.NotNil(t, err)

	// unexpected error in creation
	_, err = s.Create(ctx, CreateAppRequest{
		ID:       "error",
		Secret:   secret,
		Name:     name,
		Env:      env,
		Callback: callback,
	})
	assert.Equal(t, mock.ErrCRUD, err)

	// get
	_, err = s.Get(ctx, "none")
	assert.NotNil(t, err)
	app, err = s.Get(ctx, id)
	assert.Nil(t, err)

	// delete
	_, err = s.Delete(ctx, "none")
	assert.NotNil(t, err)
	app, err = s.Delete(ctx, id)
	assert.Nil(t, err)
	assert.Equal(t, id, app.ID)
}
