package fact

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
)

type mockService struct{}

func (m mockService) Get(ctx context.Context, connectionID int, id string) (Fact, error) {
	if id == "not_found_id" {
		return Fact{}, errors.New("not found")
	}
	rid := "rid"
	return Fact{
		Fact: entity.Fact{
			Fact:      "fact",
			Status:    "status",
			Source:    "source",
			Body:      "body",
			ISS:       "iss",
			CID:       "cid",
			RequestID: &rid,
			ID:        "id",
		},
		Attestations: []entity.Attestation{},
	}, nil
}
func (m mockService) Query(ctx context.Context, conn int, source, fact string, offset, limit int) ([]Fact, error) {
	if source == "invalid_query" {
		return []Fact{}, errors.New("expected error")
	}
	return []Fact{}, nil
}
func (m mockService) Count(ctx context.Context, conn int, source, fact string) (int, error) {
	if source == "invalid" {
		return 0, errors.New("expected count error")
	}
	return 0, nil
}
func (m mockService) Create(ctx context.Context, appID, selfID string, connection int, input CreateFactRequest) error {
	if selfID == "error" {
		return errors.New("error!")
	}
	return nil
}
func (m mockService) Update(ctx context.Context, connID int, id string, input UpdateFactRequest) (Fact, error) {
	return Fact{}, nil
}
func (m mockService) Delete(ctx context.Context, connID int, id string) error {
	if id == "error" {
		return errors.New("error!")
	}
	return nil
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
	conns := []connection.Connection{}
	if appid == "query_error" {
		return []connection.Connection{}, errors.New("expected_error_query")
	}
	conns = append(conns, connection.Connection{
		entity.Connection{
			SelfID: "selfid",
			AppID:  appid,
			Name:   "name",
		},
	})
	return conns, nil
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

func TestGetFactAPIEndpointAsPlainWithPermissions(t *testing.T) {
	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)

	rg := router.Group("/apps")
	rg.Use(acl.AuthAsPlainMiddleware([]string{"GET /apps/app_id/connections/conn_id/facts/fact_id"}))
	rg.Use(acl.NewMiddleware(filter.NewChecker()).TokenAndAccessCheckMiddleware)
	RegisterHandlers(rg, mockService{}, mockConnectionService{}, logger)

	tests := []test.APITestCase{
		{
			Name:         "success",
			Method:       "GET",
			URL:          "/apps/app_id/connections/conn_id/facts/fact_id",
			Body:         ``,
			Header:       nil,
			WantStatus:   http.StatusOK,
			WantResponse: `{"created_at":"0001-01-01T00:00:00Z", "iss":"iss", "key":"fact", "source":"source", "values":[]}			`,
		},
		{
			Name:         "connection not found",
			Method:       "GET",
			URL:          "/apps/app_id/connections/not_found_id/facts/fact_id",
			Body:         ``,
			Header:       nil,
			WantStatus:   http.StatusNotFound,
			WantResponse: `{"status":404,"error":"Not found","details":"The requested resource does not exist, or you don't have permissions to access it"}`,
		},
		{
			Name:         "connection not found",
			Method:       "GET",
			URL:          "/apps/app_id/connections/conn_id/facts/not_found_id",
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

func TestGetFactAPIEndpointAsPlainWithoutPermissions(t *testing.T) {
	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)

	rg := router.Group("/apps")
	rg.Use(acl.AuthAsPlainMiddleware([]string{}))
	rg.Use(acl.NewMiddleware(filter.NewChecker()).TokenAndAccessCheckMiddleware)
	RegisterHandlers(rg, mockService{}, mockConnectionService{}, logger)

	tests := []test.APITestCase{
		{
			Name:         "success",
			Method:       "GET",
			URL:          "/apps/app_id/connections/conn_id/facts/fact_id",
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

func TestListFactsAPIEndpointAsAdmin(t *testing.T) {
	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)

	rg := router.Group("/apps")
	rg.Use(acl.AuthAsAdminMiddleware())
	rg.Use(acl.NewMiddleware(filter.NewChecker()).TokenAndAccessCheckMiddleware)
	RegisterHandlers(rg, mockService{}, mockConnectionService{}, logger)

	tests := []test.APITestCase{
		{
			Name:         "success",
			Method:       "GET",
			URL:          "/apps/app_id/connections/conn_id/facts",
			Body:         ``,
			Header:       nil,
			WantStatus:   http.StatusOK,
			WantResponse: `{"items":[], "page":1, "page_count":0, "per_page":100, "total_count":0}`,
		},
		{
			Name:         "invalid_connection",
			Method:       "GET",
			URL:          "/apps/app_id/connections/not_found_id/facts",
			Body:         ``,
			Header:       nil,
			WantStatus:   http.StatusNotFound,
			WantResponse: `{"status":404,"error":"Not found","details":"The requested resource does not exist, or you don't have permissions to access it"}`,
		},
		{
			Name:         "internal error on count",
			Method:       "GET",
			URL:          "/apps/app_id/connections/conn_id/facts?source=invalid",
			Body:         ``,
			Header:       nil,
			WantStatus:   http.StatusInternalServerError,
			WantResponse: `There was a problem with your request. *`,
		},
		{
			Name:         "internal error on query",
			Method:       "GET",
			URL:          "/apps/app_id/connections/conn_id/facts?source=invalid_query",
			Body:         ``,
			Header:       nil,
			WantStatus:   http.StatusInternalServerError,
			WantResponse: `There was a problem with your request. *`,
		},
	}
	for _, tc := range tests {
		test.Endpoint(t, router, tc)
	}
}

func TestListFactsAPIEndpointAsPlain(t *testing.T) {
	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)

	rg := router.Group("/apps")
	rg.Use(acl.AuthAsPlainMiddleware([]string{}))
	rg.Use(acl.NewMiddleware(filter.NewChecker()).TokenAndAccessCheckMiddleware)
	RegisterHandlers(rg, mockService{}, mockConnectionService{}, logger)

	tests := []test.APITestCase{
		{
			Name:         "not found",
			Method:       "GET",
			URL:          "/apps/app_id/connections/conn_id/facts",
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

func TestCreateFactAPIEndpointAsPlainWithoutPermissions(t *testing.T) {
	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)

	rg := router.Group("/apps")
	rg.Use(acl.AuthAsPlainMiddleware([]string{}))
	rg.Use(acl.NewMiddleware(filter.NewChecker()).TokenAndAccessCheckMiddleware)
	RegisterHandlers(rg, mockService{}, mockConnectionService{}, logger)

	tests := []test.APITestCase{
		{
			Name:         "not found",
			Method:       "POST",
			URL:          "/apps/app_id/connections/conn_id/facts",
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

func TestCreateFactAPIEndpoint(t *testing.T) {
	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)

	rg := router.Group("/apps")
	rg.Use(acl.AuthAsAdminMiddleware())
	rg.Use(acl.NewMiddleware(filter.NewChecker()).TokenAndAccessCheckMiddleware)
	RegisterHandlers(rg, mockService{}, mockConnectionService{}, logger)

	tests := []test.APITestCase{
		{
			Name:         "success",
			Method:       "POST",
			URL:          "/apps/app_id/connections/conn_id/facts",
			Body:         `{"facts":[{"key": "key", "value":"value"}]}`,
			Header:       nil,
			WantStatus:   http.StatusAccepted,
			WantResponse: ``,
		},
		{
			Name:         "invalid input",
			Method:       "POST",
			URL:          "/apps/app_id/connections/conn_id/facts",
			Body:         `{]`,
			Header:       nil,
			WantStatus:   http.StatusBadRequest,
			WantResponse: `{"details":"The provided body is not valid", "error":"Invalid input", "status":400}`,
		},
		{
			Name:         "validation error",
			Method:       "POST",
			URL:          "/apps/app_id/connections/conn_id/facts",
			Body:         `{}`,
			Header:       nil,
			WantStatus:   http.StatusBadRequest,
			WantResponse: `{"details":"You should provide at least a fact to be issued", "error":"Invalid input", "status":400}`,
		},
		{
			Name:         "connection get error",
			Method:       "POST",
			URL:          "/apps/app_id/connections/not_found_id/facts",
			Body:         `{"facts":[{"key": "key", "value":"value"}]}`,
			Header:       nil,
			WantStatus:   http.StatusNotFound,
			WantResponse: `{"status":404,"error":"Not found","details":"The requested resource does not exist, or you don't have permissions to access it"}`,
		},
		{
			Name:         "creation error",
			Method:       "POST",
			URL:          "/apps/app_id/connections/error/facts",
			Body:         `{"facts":[{"key": "key", "value":"value"}]}`,
			Header:       nil,
			WantStatus:   http.StatusInternalServerError,
			WantResponse: `There was a problem with your request. *`,
		},
	}
	for _, tc := range tests {
		test.Endpoint(t, router, tc)
	}
}

func TestDeleteFactAPIEndpointAsPlainWithoutPermissions(t *testing.T) {
	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)

	rg := router.Group("/apps")
	rg.Use(acl.AuthAsPlainMiddleware([]string{}))
	rg.Use(acl.NewMiddleware(filter.NewChecker()).TokenAndAccessCheckMiddleware)
	RegisterHandlers(rg, mockService{}, mockConnectionService{}, logger)

	tests := []test.APITestCase{
		{
			Name:         "not found",
			Method:       "DELETE",
			URL:          "/apps/app_id/connections/conn_id/facts/id",
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

func TestDeleteFactAPIEndpoint(t *testing.T) {
	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)

	rg := router.Group("/apps")
	rg.Use(acl.AuthAsAdminMiddleware())
	rg.Use(acl.NewMiddleware(filter.NewChecker()).TokenAndAccessCheckMiddleware)
	RegisterHandlers(rg, mockService{}, mockConnectionService{}, logger)

	tests := []test.APITestCase{
		{
			Name:         "success",
			Method:       "DELETE",
			URL:          "/apps/app_id/connections/conn_id/facts/id",
			Body:         `{"name":"new_name"}`,
			Header:       nil,
			WantStatus:   http.StatusNoContent,
			WantResponse: ``,
		},
		{
			Name:         "modification error",
			Method:       "DELETE",
			URL:          "/apps/app_id/connections/not_found_id/facts/id",
			Body:         `{"name":"controlled_error"}`,
			Header:       nil,
			WantStatus:   http.StatusNotFound,
			WantResponse: `{"status":404,"error":"Not found","details":"The requested resource does not exist, or you don't have permissions to access it"}`,
		},
		{
			Name:         "modification error",
			Method:       "DELETE",
			URL:          "/apps/app_id/connections/conn_id/facts/error",
			Body:         `{"name":"controlled_error"}`,
			Header:       nil,
			WantStatus:   http.StatusNotFound,
			WantResponse: `{"status":404,"error":"Not found","details":"The requested resource does not exist, or you don't have permissions to access it"}`,
		},
	}
	for _, tc := range tests {
		test.Endpoint(t, router, tc)
	}
}
