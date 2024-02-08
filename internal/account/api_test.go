package account

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/joinself/restful-client/internal/entity"
	"github.com/joinself/restful-client/internal/test"
	"github.com/joinself/restful-client/pkg/acl"
	"github.com/joinself/restful-client/pkg/log"
)

const (
	ErrorUsername = "throw_error"
	LongString    = "1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890"
)

type mockService struct{}

func (m mockService) Get(ctx context.Context, username, password string) (Account, error) {
	return Account{}, nil
}

func (m mockService) Create(ctx context.Context, input CreateAccountRequest) (Account, error) {
	if input.Username == ErrorUsername {
		return Account{}, errors.New("expected error")
	}
	return Account{
		entity.Account{
			UserName:               input.Username,
			Resources:              strings.Join(input.Resources, ","),
			RequiresPasswordChange: 0,
		},
	}, nil
}

func (m mockService) SetPassword(ctx context.Context, username, password, newPassword string) error {
	if newPassword == ErrorUsername {
		return errors.New("expected error")
	}

	return nil
}

func (m mockService) Delete(ctx context.Context, username string) error {
	println(".--------> " + username)
	if username == "errored" {
		return errors.New("username not found")
	}
	return nil
}

func (m mockService) Count(ctx context.Context) (int, error) {
	return 0, nil
}

func TestCreateAccountAPIEndpointAsAdmin(t *testing.T) {
	var validPwd = "valid_password"
	var validUsr = "valid_user"

	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)

	rg := router.Group("/accounts")
	rg.Use(acl.AuthAsAdminMiddleware())
	RegisterHandlers(rg, mockService{}, logger)

	tests := []test.APITestCase{
		{
			Name:         "success",
			Method:       "POST",
			URL:          "/accounts",
			Body:         `{"username":"test_larger","password":"pass_larger","resources":[]}`,
			Header:       nil,
			WantStatus:   http.StatusCreated,
			WantResponse: `{"requires_password_change":0, "resources":"", "user_name":"test_larger"}`,
		},
		{
			Name:         "invalid input",
			Method:       "POST",
			URL:          "/accounts",
			Body:         `{]`,
			Header:       nil,
			WantStatus:   http.StatusBadRequest,
			WantResponse: `{"details":"The provided body is not valid", "error":"Invalid input", "status":400}`,
		},
		{
			Name:         "required input",
			Method:       "POST",
			URL:          "/accounts",
			Body:         `{}`,
			Header:       nil,
			WantStatus:   http.StatusBadRequest,
			WantResponse: `{"details":"password: cannot be blank; username: cannot be blank.", "error":"Invalid input", "status":400}`,
		},
		{
			Name:         "username too long",
			Method:       "POST",
			URL:          "/accounts",
			Body:         fmt.Sprintf(`{"username":"%s","password":"%s"}`, LongString, validPwd),
			Header:       nil,
			WantStatus:   http.StatusBadRequest,
			WantResponse: `{"details":"username: the length must be between 5 and 128.", "error":"Invalid input", "status":400}`,
		},
		{
			Name:         "password too long",
			Method:       "POST",
			URL:          "/accounts",
			Body:         fmt.Sprintf(`{"username":"%s","password":"%s"}`, validUsr, LongString),
			Header:       nil,
			WantStatus:   http.StatusBadRequest,
			WantResponse: `{"details":"password: the length must be between 5 and 128.", "error":"Invalid input", "status":400}`,
		},
		{
			Name:         "username too short",
			Method:       "POST",
			URL:          "/accounts",
			Body:         fmt.Sprintf(`{"username":"%s","password":"%s"}`, "a", validPwd),
			Header:       nil,
			WantStatus:   http.StatusBadRequest,
			WantResponse: `{"details":"username: the length must be between 5 and 128.", "error":"Invalid input", "status":400}`,
		},
		{
			Name:         "password too long",
			Method:       "POST",
			URL:          "/accounts",
			Body:         fmt.Sprintf(`{"username":"%s","password":"%s"}`, validUsr, "a"),
			Header:       nil,
			WantStatus:   http.StatusBadRequest,
			WantResponse: `{"details":"password: the length must be between 5 and 128.", "error":"Invalid input", "status":400}`,
		},
		{
			Name:         "username blank",
			Method:       "POST",
			URL:          "/accounts",
			Body:         fmt.Sprintf(`{"username":"%s","password":"%s"}`, "", validPwd),
			Header:       nil,
			WantStatus:   http.StatusBadRequest,
			WantResponse: `{"details":"username: cannot be blank.", "error":"Invalid input", "status":400}`,
		},
		{
			Name:         "password blank",
			Method:       "POST",
			URL:          "/accounts",
			Body:         fmt.Sprintf(`{"username":"%s","password":"%s"}`, validUsr, ""),
			Header:       nil,
			WantStatus:   http.StatusBadRequest,
			WantResponse: `{"details":"password: cannot be blank.", "error":"Invalid input", "status":400}`,
		},
		{
			Name:         "error on creation",
			Method:       "POST",
			URL:          "/accounts",
			Body:         fmt.Sprintf(`{"username":"%s","password":"%s"}`, ErrorUsername, validPwd),
			Header:       nil,
			WantStatus:   http.StatusInternalServerError,
			WantResponse: `There was a problem with your request. *`,
		},
	}
	for _, tc := range tests {
		test.Endpoint(t, router, tc)
	}
}

func TestCreateAccountAPIEndpointAsPlain(t *testing.T) {
	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)

	rg := router.Group("/accounts")
	rg.Use(acl.AuthAsPlainMiddleware([]string{}))
	RegisterHandlers(rg, mockService{}, logger)

	tests := []test.APITestCase{
		{
			Name:         "success",
			Method:       "POST",
			URL:          "/accounts",
			Body:         `{"username":"test_larger","password":"pass_larger","resources":[]}`,
			Header:       nil,
			WantStatus:   http.StatusNotFound,
			WantResponse: `{"status":404,"error":"Not found","details":"The requested resource does not exist, or you don't have permissions to access it"}`,
		},
	}
	for _, tc := range tests {
		test.Endpoint(t, router, tc)
	}
}

func TestDeleteAccountAPIEndpointAsAdmin(t *testing.T) {
	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)

	rg := router.Group("/accounts")
	rg.Use(acl.AuthAsAdminMiddleware())
	RegisterHandlers(rg, mockService{}, logger)

	tests := []test.APITestCase{
		{
			Name:         "success",
			Method:       "DELETE",
			URL:          "/accounts/username",
			Body:         ``,
			Header:       nil,
			WantStatus:   http.StatusNoContent,
			WantResponse: ``,
		},
		{
			Name:         "error deleting",
			Method:       "DELETE",
			URL:          "/accounts/errored",
			Body:         ``,
			Header:       nil,
			WantStatus:   http.StatusNotFound,
			WantResponse: `{"status":404,"error":"Not found","details":"The requested resource does not exist, or you don't have permissions to access it"}`,
		},
	}
	for _, tc := range tests {
		test.Endpoint(t, router, tc)
	}
}

func TestDeleteAccountAPIEndpointAsPlain(t *testing.T) {
	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)

	rg := router.Group("/accounts")
	rg.Use(acl.AuthAsPlainMiddleware([]string{}))
	RegisterHandlers(rg, mockService{}, logger)

	tests := []test.APITestCase{
		{
			Name:         "success",
			Method:       "DELETE",
			URL:          "/accounts/username",
			Body:         ``,
			Header:       nil,
			WantStatus:   http.StatusNotFound,
			WantResponse: `{"status":404,"error":"Not found","details":"The requested resource does not exist, or you don't have permissions to access it"}`,
		},
	}
	for _, tc := range tests {
		test.Endpoint(t, router, tc)
	}
}

/*
	func OTestChangePasswordAPIEndpointAsAdmin(t *testing.T) {
		var longString = "1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890"
		var validPwd = "valid_password"
		var validUsr = "valid_user"

		logger, _ := log.NewForTest()
		router := test.MockRouter(logger)

		rg := router.Group("/accounts")
		rg.Use(acl.AuthAsAdminMiddleware())
		RegisterHandlers(rg, mockService{}, logger)

		tests := []test.APITestCase{
			{
				Name:         "success",
				Method:       "PUT",
				URL:          "/accounts/username/password",
				Body:         `{"new_password":"new_password","password":"pass_larger","resources":[]}`,
				Header:       nil,
				WantStatus:   http.StatusCreated,
				WantResponse: `{"requires_password_change":0, "resources":"", "user_name":"test_larger"}`,
			},
			{
				Name:         "invalid input",
				Method:       "PUT",
				URL:          "/accounts/username/password",
				Body:         `{]`,
				Header:       nil,
				WantStatus:   http.StatusBadRequest,
				WantResponse: `{"details":"The provided body is not valid", "error":"Invalid input", "status":400}`,
			},
			{
				Name:         "required input",
				Method:       "PUT",
				URL:          "/accounts/username/password",
				Body:         `{}`,
				Header:       nil,
				WantStatus:   http.StatusBadRequest,
				WantResponse: `{"details":"password: cannot be blank; username: cannot be blank.", "error":"Invalid input", "status":400}`,
			},
			{
				Name:         "username too long",
				Method:       "PUT",
				URL:          "/accounts/username/password",
				Body:         fmt.Sprintf(`{"username":"%s","password":"%s"}`, longString, validPwd),
				Header:       nil,
				WantStatus:   http.StatusBadRequest,
				WantResponse: `{"details":"username: the length must be between 5 and 128.", "error":"Invalid input", "status":400}`,
			},
			{
				Name:         "password too long",
				Method:       "PUT",
				URL:          "/accounts/username/password",
				Body:         fmt.Sprintf(`{"username":"%s","password":"%s"}`, validUsr, longString),
				Header:       nil,
				WantStatus:   http.StatusBadRequest,
				WantResponse: `{"details":"password: the length must be between 5 and 128.", "error":"Invalid input", "status":400}`,
			},
			{
				Name:         "username too short",
				Method:       "PUT",
				URL:          "/accounts/username/password",
				Body:         fmt.Sprintf(`{"username":"%s","password":"%s"}`, "a", validPwd),
				Header:       nil,
				WantStatus:   http.StatusBadRequest,
				WantResponse: `{"details":"username: the length must be between 5 and 128.", "error":"Invalid input", "status":400}`,
			},
			{
				Name:         "password too long",
				Method:       "PUT",
				URL:          "/accounts/username/password",
				Body:         fmt.Sprintf(`{"username":"%s","password":"%s"}`, validUsr, "a"),
				Header:       nil,
				WantStatus:   http.StatusBadRequest,
				WantResponse: `{"details":"password: the length must be between 5 and 128.", "error":"Invalid input", "status":400}`,
			},
			{
				Name:         "username blank",
				Method:       "PUT",
				URL:          "/accounts/username/password",
				Body:         fmt.Sprintf(`{"username":"%s","password":"%s"}`, "", validPwd),
				Header:       nil,
				WantStatus:   http.StatusBadRequest,
				WantResponse: `{"details":"username: cannot be blank.", "error":"Invalid input", "status":400}`,
			},
			{
				Name:         "password blank",
				Method:       "PUT",
				URL:          "/accounts/username/password",
				Body:         fmt.Sprintf(`{"username":"%s","password":"%s"}`, validUsr, ""),
				Header:       nil,
				WantStatus:   http.StatusBadRequest,
				WantResponse: `{"details":"password: cannot be blank.", "error":"Invalid input", "status":400}`,
			},
			{
				Name:         "error on creation",
				Method:       "PUT",
				URL:          "/accounts/username/password",
				Body:         fmt.Sprintf(`{"username":"%s","password":"%s"}`, ErrorUsername, validPwd),
				Header:       nil,
				WantStatus:   http.StatusInternalServerError,
				WantResponse: `There was a problem with your request. *`,
			},
		}
		for _, tc := range tests {
			test.Endpoint(t, router, tc)
		}
	}
*/
func TestChangePasswordAPIEndpointAsPlain(t *testing.T) {
	var validPwd = "password"

	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)

	rg := router.Group("/accounts")
	rg.Use(acl.AuthAsPlainMiddleware([]string{}))
	RegisterHandlers(rg, mockService{}, logger)

	tests := []test.APITestCase{
		{
			Name:         "success",
			Method:       "PUT",
			URL:          "/accounts/john/password",
			Body:         `{"new_password":"new_password","password":"old_password"}`,
			Header:       nil,
			WantStatus:   http.StatusOK,
			WantResponse: ``,
		},
		{
			Name:         "invalid input",
			Method:       "PUT",
			URL:          "/accounts/john/password",
			Body:         `{]`,
			Header:       nil,
			WantStatus:   http.StatusBadRequest,
			WantResponse: `{"details":"The provided body is not valid", "error":"Invalid input", "status":400}`,
		},
		{
			Name:         "required input",
			Method:       "PUT",
			URL:          "/accounts/john/password",
			Body:         `{}`,
			Header:       nil,
			WantStatus:   http.StatusBadRequest,
			WantResponse: `{"details":"new_password: cannot be blank; password: cannot be blank.", "error":"Invalid input", "status":400}`,
		},
		{
			Name:         "new password too long",
			Method:       "PUT",
			URL:          "/accounts/john/password",
			Body:         fmt.Sprintf(`{"new_password":"%s","password":"%s"}`, LongString, validPwd),
			Header:       nil,
			WantStatus:   http.StatusBadRequest,
			WantResponse: `{"details":"new_password: the length must be between 5 and 128.", "error":"Invalid input", "status":400}`,
		},
		{
			Name:         "new password too short",
			Method:       "PUT",
			URL:          "/accounts/john/password",
			Body:         fmt.Sprintf(`{"new_password":"%s","password":"%s"}`, "o", validPwd),
			Header:       nil,
			WantStatus:   http.StatusBadRequest,
			WantResponse: `{"details":"new_password: the length must be between 5 and 128.", "error":"Invalid input", "status":400}`,
		},
		{
			Name:         "new password blank",
			Method:       "PUT",
			URL:          "/accounts/john/password",
			Body:         fmt.Sprintf(`{"new_password":"%s","password":"%s"}`, "", validPwd),
			Header:       nil,
			WantStatus:   http.StatusBadRequest,
			WantResponse: `{"details":"new_password: cannot be blank.", "error":"Invalid input", "status":400}`,
		},
		{
			Name:         "error on creation",
			Method:       "PUT",
			URL:          "/accounts/john/password",
			Body:         fmt.Sprintf(`{"new_password":"%s","password":"%s"}`, ErrorUsername, validPwd),
			Header:       nil,
			WantStatus:   http.StatusInternalServerError,
			WantResponse: `There was a problem with your request. *`,
		},
		{
			Name:         "not matching username",
			Method:       "PUT",
			URL:          "/accounts/bob/password",
			Body:         `{"new_password":"new_password","password":"old_password"}`,
			Header:       nil,
			WantStatus:   http.StatusNotFound,
			WantResponse: `{"details":"The requested resource does not exist, or you don't have permissions to access it", "error":"Not found", "status":404}`,
		},
	}
	for _, tc := range tests {
		test.Endpoint(t, router, tc)
	}
}

func TestChangePasswordAPIEndpointPublic(t *testing.T) {
	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)

	rg := router.Group("/accounts")
	RegisterHandlers(rg, mockService{}, logger)

	tests := []test.APITestCase{
		{
			Name:         "success",
			Method:       "PUT",
			URL:          "/accounts/john/password",
			Body:         `{"new_password":"new_password","password":"old_password"}`,
			Header:       nil,
			WantStatus:   http.StatusNotFound,
			WantResponse: `{"details":"The requested resource does not exist, or you don't have permissions to access it", "error":"Not found", "status":404}`,
		},
	}
	for _, tc := range tests {
		test.Endpoint(t, router, tc)
	}
}
