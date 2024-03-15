package request

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/joinself/restful-client/internal/connection"
	"github.com/joinself/restful-client/internal/entity"
	"github.com/joinself/restful-client/internal/test"
	"github.com/joinself/restful-client/pkg/acl"
	"github.com/joinself/restful-client/pkg/filter"
	"github.com/joinself/restful-client/pkg/log"
	"github.com/joinself/restful-client/pkg/support"

	selffact "github.com/joinself/self-go-sdk/fact"
)

type mockService struct{}

func (m mockService) Get(ctx context.Context, appID, id string) (ExtRequest, error) {
	if id == "not_found_id" {
		return ExtRequest{}, errors.New("not found")
	}

	return ExtRequest{}, nil
}

func (m mockService) Create(ctx context.Context, appID string, conn *entity.Connection, input CreateRequest) (ExtRequest, error) {
	if appID == "error" {
		return ExtRequest{}, errors.New("error!")
	}
	return ExtRequest{}, nil
}

func (m mockService) CreateFactsFromResponse(conn entity.Connection, req entity.Request, facts []selffact.Fact) []entity.Fact {
	return []entity.Fact{}
}

func (m mockService) SetRunner(runner support.SelfClientGetter) {
	return
}

type mockConnectionService struct{}

func (m mockConnectionService) Get(ctx context.Context, appid, selfid string) (connection.Connection, error) {
	if selfid == "not_found_id" {
		return connection.Connection{}, errors.New("expected not found")
	}
	return connection.Connection{
		entity.Connection{
			SelfID: "selfid",
			AppID:  appid,
			Name:   "name",
		},
	}, nil
}

func (m mockConnectionService) Query(ctx context.Context, appid string, offset, limit int) ([]connection.Connection, error) {
	return []connection.Connection{}, nil
}

func (m mockConnectionService) Count(ctx context.Context, appid string) (int, error) {
	if appid == "count_error" {
		return 0, errors.New("expected_error_count")
	}
	return 1, nil
}

func (m mockConnectionService) Create(ctx context.Context, appid string, input connection.CreateConnectionRequest) (connection.Connection, error) {
	if input.SelfID == "controlled_error" {
		return connection.Connection{}, errors.New("controlled error")
	}
	return connection.Connection{}, nil
}

func (m mockConnectionService) Update(ctx context.Context, appid, selfid string, input connection.UpdateConnectionRequest) (connection.Connection, error) {
	if input.Name == "controlled_error" {
		return connection.Connection{}, errors.New("controlled error")
	}
	return connection.Connection{}, nil
}

func (m mockConnectionService) Delete(ctx context.Context, appid, selfid string) (connection.Connection, error) {
	if selfid == "controlled_error" {
		return connection.Connection{}, errors.New("controlled error")
	}
	return connection.Connection{}, nil
}

func TestGetRequestAPIEndpointAsPlainWithPermissions(t *testing.T) {
	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)

	rg := router.Group("/apps")
	rg.Use(acl.AuthAsPlainMiddleware([]string{"GET /apps/app_id/requests/request_jti"}))
	rg.Use(acl.NewMiddleware(filter.NewChecker()).Process)
	RegisterHandlers(rg, mockService{}, mockConnectionService{}, logger)

	tests := []test.APITestCase{
		{
			Name:         "success",
			Method:       "GET",
			URL:          "/apps/app_id/requests/request_jti",
			Body:         ``,
			Header:       nil,
			WantStatus:   http.StatusOK,
			WantResponse: `{"app_id":"app_id", "id":""}`,
		},
		{
			Name:         "connection not found",
			Method:       "GET",
			URL:          "/apps/app_id/connections/not_found_id/requests/request_jti",
			Body:         ``,
			Header:       nil,
			WantStatus:   http.StatusNotFound,
			WantResponse: `{"status":404,"error":"Not found","details":"The requested resource does not exist, or you don't have permissions to access it"}`,
		},
		{
			Name:         "connection not found",
			Method:       "GET",
			URL:          "/apps/app_id/requests/not_found_id",
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

func TestGetRequestAPIEndpointAsPlainWithoutPermissions(t *testing.T) {
	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)

	rg := router.Group("/apps")
	rg.Use(acl.AuthAsPlainMiddleware([]string{}))
	rg.Use(acl.NewMiddleware(filter.NewChecker()).Process)
	RegisterHandlers(rg, mockService{}, mockConnectionService{}, logger)

	tests := []test.APITestCase{
		{
			Name:         "success",
			Method:       "GET",
			URL:          "/apps/app_id/requests/request_jti",
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

func TestCreateRequestAPIEndpointAsPlainWithoutPermissions(t *testing.T) {
	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)

	rg := router.Group("/apps")
	rg.Use(acl.AuthAsPlainMiddleware([]string{}))
	rg.Use(acl.NewMiddleware(filter.NewChecker()).Process)
	RegisterHandlers(rg, mockService{}, mockConnectionService{}, logger)

	tests := []test.APITestCase{
		{
			Name:         "not found",
			Method:       "POST",
			URL:          "/apps/app_id/requests",
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

func TestCreateRequestAPIEndpoint(t *testing.T) {
	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)

	rg := router.Group("/apps")
	rg.Use(acl.AuthAsAdminMiddleware())
	rg.Use(acl.NewMiddleware(filter.NewChecker()).Process)
	RegisterHandlers(rg, mockService{}, mockConnectionService{}, logger)

	tests := []test.APITestCase{
		{
			Name:         "success",
			Method:       "POST",
			URL:          "/apps/app_id/requests",
			Body:         `{"type":"auth"}`,
			Header:       nil,
			WantStatus:   http.StatusAccepted,
			WantResponse: ``,
		},
		{
			Name:         "invalid input",
			Method:       "POST",
			URL:          "/apps/app_id/requests",
			Body:         `{]`,
			Header:       nil,
			WantStatus:   http.StatusBadRequest,
			WantResponse: `{"details":"The provided body is not valid", "error":"Invalid input", "status":400}`,
		},
		{
			Name:         "validation error",
			Method:       "POST",
			URL:          "/apps/app_id/requests",
			Body:         `{}`,
			Header:       nil,
			WantStatus:   http.StatusBadRequest,
			WantResponse: `{"details":"type: cannot be blank.", "error":"Invalid input", "status":400}`,
		},
		{
			Name:         "creation error",
			Method:       "POST",
			URL:          "/apps/error/requests",
			Body:         `{"type":"auth"}`,
			Header:       nil,
			WantStatus:   http.StatusInternalServerError,
			WantResponse: `There was a problem with your request. *`,
		},
	}
	for _, tc := range tests {
		test.Endpoint(t, router, tc)
	}
}
