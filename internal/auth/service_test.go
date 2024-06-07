package auth

import (
	"context"
	"testing"
	"time"

	"github.com/joinself/restful-client/internal/config"
	"github.com/joinself/restful-client/internal/entity"
	"github.com/joinself/restful-client/internal/errors"
	"github.com/joinself/restful-client/pkg/log"
	"github.com/joinself/restful-client/pkg/mock"
	"github.com/stretchr/testify/assert"
)

func Test_service_Authenticate(t *testing.T) {
	logger, _ := log.NewForTest()
	cfg := config.Config{
		JWTSigningKey:                 "test",
		JWTExpirationTimeInHours:      100,
		RefreshTokenExpirationInHours: 100,
		User:                          "demo",
		Password:                      "pass",
	}
	s := NewService(&cfg, &mock.AccountRepositoryMock{}, &mock.AppRepositoryMock{}, logger)
	_, err := s.Login(context.Background(), "unknown", "bad")
	assert.Equal(t, errors.Unauthorized("account does not exist"), err)
	resp, err := s.Login(context.Background(), "demo", "pass")
	assert.Nil(t, err)
	assert.NotEmpty(t, resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)
}

func Test_service_authenticate(t *testing.T) {
	logger, _ := log.NewForTest()
	accountMock := mock.AccountRepositoryMock{}
	s := service{"test", 100, 100, "demo", "pass", &accountMock, &mock.AppRepositoryMock{}, logger}
	assert.Nil(t, s.authenticate(context.Background(), "unknown", "bad"))
	assert.NotNil(t, s.authenticate(context.Background(), "demo", "pass"))
	accountMock.Create(context.Background(), entity.Account{
		UserName: "foooo",
		Password: "baaar",
	})
	assert.NotNil(t, s.authenticate(context.Background(), "foooo", "baaar"))
}

func Test_service_GenerateJWT(t *testing.T) {
	logger, _ := log.NewForTest()
	s := service{"test", 100, 100, "demo", "pass", &mock.AccountRepositoryMock{}, &mock.AppRepositoryMock{}, logger}
	token, err := s.generateJWT(entity.User{
		ID:   "100",
		Name: "demo",
	})
	if assert.Nil(t, err) {
		assert.NotEmpty(t, token)
	}
}

func Test_refresh_token(t *testing.T) {
	logger, _ := log.NewForTest()
	cfg := config.Config{
		JWTSigningKey:                 "test",
		JWTExpirationTimeInHours:      100,
		RefreshTokenExpirationInHours: 100,
		User:                          "demo",
		Password:                      "pass",
	}

	accountMock := &mock.AccountRepositoryMock{}

	s := NewService(&cfg, accountMock, &mock.AppRepositoryMock{}, logger)
	// Config account
	resp, err := s.Login(context.Background(), "demo", "pass")
	assert.Nil(t, err)
	assert.NotEmpty(t, resp.RefreshToken)

	time.Sleep(time.Duration(1 * time.Second))
	resp2, err := s.Refresh(context.Background(), resp.RefreshToken)
	assert.Nil(t, err)
	assert.NotEmpty(t, resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)
	assert.NotEqual(t, resp.AccessToken, resp2.AccessToken)
	assert.NotEqual(t, resp2.AccessToken, resp2.RefreshToken)

	// DB based account
	accountMock.Create(context.Background(), entity.Account{
		UserName: "john",
		Password: "smith",
	})
	resp3, err := s.Login(context.Background(), "john", "smith")
	assert.Nil(t, err)
	assert.NotEmpty(t, resp3.RefreshToken)
	resp4, err := s.Refresh(context.Background(), resp3.RefreshToken)
	assert.Nil(t, err)
	assert.NotEqual(t, resp3.AccessToken, resp4.AccessToken)
	assert.NotEqual(t, resp4.AccessToken, resp4.RefreshToken)

	// Non existing account
	accountMock.Items = []entity.Account{}
	_, err = s.Refresh(context.Background(), resp3.RefreshToken)
	assert.Error(t, errors.Unauthorized("account not found"))
}
