package account

import (
	"context"
	"testing"

	"github.com/joinself/restful-client/pkg/log"
	"github.com/joinself/restful-client/pkg/mock"
	"github.com/stretchr/testify/assert"
)

const longString = "1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890"

func TestCreateAccountRequest_Validate(t *testing.T) {
	tests := []struct {
		name      string
		model     CreateAccountRequest
		wantError bool
	}{
		{"success", CreateAccountRequest{Username: "username", Password: "password"}, false},
		{"required", CreateAccountRequest{}, true},
		{"too long usr", CreateAccountRequest{Username: longString, Password: "password"}, true},
		{"too long pwd", CreateAccountRequest{Username: "user", Password: longString}, true},
		{"too short user", CreateAccountRequest{Username: "", Password: "password"}, true},
		{"too long pwd", CreateAccountRequest{Username: "user", Password: ""}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.model.Validate()
			assert.Equal(t, tt.wantError, err != nil)
		})
	}
}

func TestUpdateAccountRequest_Validate(t *testing.T) {
	tests := []struct {
		name      string
		model     UpdateAccountRequest
		wantError bool
	}{
		{"success", UpdateAccountRequest{NewPassword: "username", Password: "password"}, false},
		{"required", UpdateAccountRequest{}, true},
		{"too long new pwd", UpdateAccountRequest{NewPassword: longString, Password: "password"}, true},
		{"too long pwd", UpdateAccountRequest{NewPassword: "user", Password: longString}, true},
		{"too short new pwd", UpdateAccountRequest{NewPassword: "", Password: "password"}, true},
		{"too long pwd", UpdateAccountRequest{NewPassword: "user", Password: ""}, true},
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

	s := NewService(&mock.AccountRepositoryMock{}, logger)

	ctx := context.Background()

	// initial count
	count, _ := s.Count(ctx)
	assert.Equal(t, 0, count)

	// successful creation
	username := "selfid"
	account, err := s.Create(ctx, CreateAccountRequest{
		Username:  username,
		Password:  "password",
		Resources: []string{"appid"},
	})
	assert.Nil(t, err)
	assert.Equal(t, username, account.UserName)
	assert.NotEmpty(t, account.CreatedAt)
	assert.NotEmpty(t, account.UpdatedAt)
	count, _ = s.Count(ctx)
	assert.Equal(t, 1, count)

	// validation error in creation
	_, err = s.Create(ctx, CreateAccountRequest{
		Username: "",
		Password: "password",
	})
	assert.NotNil(t, err)
	count, _ = s.Count(ctx)
	assert.Equal(t, 1, count)

	// unexpected error in creation
	_, err = s.Create(ctx, CreateAccountRequest{
		Username: "error",
		Password: "password",
	})
	assert.Equal(t, mock.ErrCRUD, err)
	count, _ = s.Count(ctx)
	assert.Equal(t, 1, count)

	_, err = s.Create(ctx, CreateAccountRequest{
		Username: "test2",
		Password: "password",
	})
	assert.Nil(t, err)
	count, _ = s.Count(ctx)
	assert.Equal(t, 2, count)

	// get
	_, err = s.Get(ctx, "none", "password")
	assert.NotNil(t, err)
	account, err = s.Get(ctx, username, "password")
	assert.Nil(t, err)
	assert.Equal(t, "appid", account.Resources)
	assert.Equal(t, username, account.UserName)

	// delete
	err = s.Delete(ctx, "none")
	assert.NotNil(t, err)
	err = s.Delete(ctx, username)
	assert.Nil(t, err)
	count, _ = s.Count(ctx)
	assert.Equal(t, 1, count)
}
