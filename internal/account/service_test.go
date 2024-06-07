package account

import (
	"context"
	"testing"

	"github.com/joinself/restful-client/pkg/log"
	"github.com/joinself/restful-client/pkg/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	longString    = "1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890"
	validPassword = "Pass_larger_1!"
)

func TestCreateAccountRequest_Validate(t *testing.T) {
	tests := []struct {
		name      string
		model     CreateAccountRequest
		wantError bool
	}{
		{"success", CreateAccountRequest{Username: "username", Password: validPassword}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.model.Validate()
			assert.Equal(t, tt.wantError, err != nil)
		})
	}
}

func TestChangePasswordRequest_Validate(t *testing.T) {
	tests := []struct {
		name      string
		model     ChangePasswordRequest
		wantError bool
	}{
		{"success", ChangePasswordRequest{NewPassword: validPassword, Password: validPassword}, false},
		{"required", ChangePasswordRequest{}, true},
		{"too long new pwd", ChangePasswordRequest{NewPassword: longString, Password: validPassword}, true},
		{"too long pwd", ChangePasswordRequest{NewPassword: "user", Password: longString}, true},
		{"too short new pwd", ChangePasswordRequest{NewPassword: "", Password: validPassword}, true},
		{"too long pwd", ChangePasswordRequest{NewPassword: "user", Password: ""}, true},
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
		Password:  validPassword,
		Resources: []string{"appid"},
	})
	assert.Nil(t, err)
	assert.Equal(t, username, account.UserName)
	assert.Equal(t, account.RequiresPasswordChange, 1)
	assert.NotEmpty(t, account.CreatedAt)
	assert.NotEmpty(t, account.UpdatedAt)
	count, _ = s.Count(ctx)
	assert.Equal(t, 1, count)

	// get
	_, err = s.Get(ctx, "none", validPassword)
	assert.NotNil(t, err)
	account, err = s.Get(ctx, username, validPassword)
	assert.Nil(t, err)
	rr := account.GetResources()
	require.Equal(t, len(rr), 1)
	assert.Equal(t, "appid", rr[0])
	assert.Equal(t, username, account.UserName)

	// delete
	err = s.Delete(ctx, "none")
	assert.NotNil(t, err)
	err = s.Delete(ctx, username)
	assert.Nil(t, err)
	count, _ = s.Count(ctx)
	assert.Equal(t, 0, count)

	// successful creation without password change
	username2 := "selfid2"
	f := false
	account2, err := s.Create(ctx, CreateAccountRequest{
		Username:               username2,
		Password:               validPassword,
		Resources:              []string{"appid"},
		RequiresPasswordChange: &f,
	})
	assert.Nil(t, err)
	assert.Equal(t, username2, account2.UserName)
	assert.Equal(t, account2.RequiresPasswordChange, 0)
	assert.NotEmpty(t, account2.CreatedAt)
	assert.NotEmpty(t, account2.UpdatedAt)

}
