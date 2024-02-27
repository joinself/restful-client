package message

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

func (m mockService) Get(ctx context.Context, connectionID int, jti string) (Message, error) {
	if jti == "not_found_id" {
		return Message{}, errors.New("not found")
	}

	return Message{
		Message: entity.Message{
			Body: "body",
			ISS:  "iss",
			CID:  "cid",
		},
	}, nil
}

func (m mockService) Query(ctx context.Context, connection int, messagesSince int, offset, limit int) ([]Message, error) {
	if messagesSince == 98 {
		return []Message{}, errors.New("expected error")
	}
	return []Message{}, nil
}
func (m mockService) Count(ctx context.Context, connectionID, messagesSince int) (int, error) {
	if messagesSince == 99 {
		return 0, errors.New("expected count error")
	}
	return 0, nil
}
func (m mockService) Create(ctx context.Context, appID, connectionID string, connection int, input CreateMessageRequest) (Message, error) {
	if input.Body == "error" {
		return Message{}, errors.New("error!")
	}
	return Message{}, nil
}
func (m mockService) Update(ctx context.Context, appID string, connectionID int, selfID string, jti string, req UpdateMessageRequest) (Message, error) {
	if req.Body == "error" {
		return Message{}, errors.New("error!")
	}
	return Message{}, nil

}
func (m mockService) Delete(ctx context.Context, connectionID int, jti string) error {
	if jti == "error" {
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

func TestGetMessageAPIEndpointAsPlainWithPermissions(t *testing.T) {
	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)

	rg := router.Group("/apps")
	rg.Use(acl.AuthAsPlainMiddleware([]string{"GET /apps/app_id/*"}))
	rg.Use(acl.NewMiddleware(filter.NewChecker()).Process)
	RegisterHandlers(rg, mockService{}, mockConnectionService{}, logger)

	tests := []test.APITestCase{
		{
			Name:         "success",
			Method:       "GET",
			URL:          "/apps/app_id/connections/conn_id/messages/message_jti",
			Body:         ``,
			Header:       nil,
			WantStatus:   http.StatusOK,
			WantResponse: `{"body":"body", "cid":"cid", "created_at":"0001-01-01T00:00:00Z", "iat":"0001-01-01T00:00:00Z", "iss":"iss", "jti":"", "rid":"", "updated_at":"0001-01-01T00:00:00Z"}`,
		},
		{
			Name:         "connection not found",
			Method:       "GET",
			URL:          "/apps/app_id/connections/not_found_id/messages/message_jti",
			Body:         ``,
			Header:       nil,
			WantStatus:   http.StatusNotFound,
			WantResponse: `{"status":404,"error":"Not found","details":"The requested resource does not exist, or you don't have permissions to access it"}`,
		},
		{
			Name:         "connection not found",
			Method:       "GET",
			URL:          "/apps/app_id/connections/conn_id/messages/not_found_id",
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

func TestGetMessageAPIEndpointAsPlainWithoutPermissions(t *testing.T) {
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
			URL:          "/apps/app_id/connections/conn_id/messages/message_jti",
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

func TestListMessagesAPIEndpointAsAdmin(t *testing.T) {
	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)

	rg := router.Group("/apps")
	rg.Use(acl.AuthAsAdminMiddleware())
	rg.Use(acl.NewMiddleware(filter.NewChecker()).Process)
	RegisterHandlers(rg, mockService{}, mockConnectionService{}, logger)

	tests := []test.APITestCase{
		{
			Name:         "success",
			Method:       "GET",
			URL:          "/apps/app_id/connections/conn_id/messages",
			Body:         ``,
			Header:       nil,
			WantStatus:   http.StatusOK,
			WantResponse: `{"items":[], "page":1, "page_count":0, "per_page":100, "total_count":0}`,
		},
		{
			Name:         "invalid_connection",
			Method:       "GET",
			URL:          "/apps/app_id/connections/not_found_id/messages",
			Body:         ``,
			Header:       nil,
			WantStatus:   http.StatusNotFound,
			WantResponse: `{"status":404,"error":"Not found","details":"The requested resource does not exist, or you don't have permissions to access it"}`,
		},
		{
			Name:         "internal error on count",
			Method:       "GET",
			URL:          "/apps/app_id/connections/conn_id/messages?messages_since=99",
			Body:         ``,
			Header:       nil,
			WantStatus:   http.StatusInternalServerError,
			WantResponse: `There was a problem with your request. *`,
		},
		{
			Name:         "internal error on query",
			Method:       "GET",
			URL:          "/apps/app_id/connections/conn_id/messages?messages_since=98",
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

func TestListMessagesAPIEndpointAsPlain(t *testing.T) {
	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)

	rg := router.Group("/apps")
	rg.Use(acl.AuthAsPlainMiddleware([]string{}))
	rg.Use(acl.NewMiddleware(filter.NewChecker()).Process)
	RegisterHandlers(rg, mockService{}, mockConnectionService{}, logger)

	tests := []test.APITestCase{
		{
			Name:         "not found",
			Method:       "GET",
			URL:          "/apps/app_id/connections/conn_id/messages",
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

func TestCreateMessageAPIEndpointAsPlainWithoutPermissions(t *testing.T) {
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
			URL:          "/apps/app_id/connections/conn_id/messages",
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

func TestCreateMessageAPIEndpoint(t *testing.T) {
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
			URL:          "/apps/app_id/connections/conn_id/messages",
			Body:         `{"body":"hello"}`,
			Header:       nil,
			WantStatus:   http.StatusAccepted,
			WantResponse: ``,
		},
		{
			Name:         "invalid input",
			Method:       "POST",
			URL:          "/apps/app_id/connections/conn_id/messages",
			Body:         `{]`,
			Header:       nil,
			WantStatus:   http.StatusBadRequest,
			WantResponse: `{"details":"The provided body is not valid", "error":"Invalid input", "status":400}`,
		},
		{
			Name:         "validation error",
			Method:       "POST",
			URL:          "/apps/app_id/connections/conn_id/messages",
			Body:         `{}`,
			Header:       nil,
			WantStatus:   http.StatusBadRequest,
			WantResponse: `{"details":"body: cannot be blank.", "error":"Invalid input", "status":400}`,
		},
		{
			Name:         "connection get error",
			Method:       "POST",
			URL:          "/apps/app_id/connections/not_found_id/messages",
			Body:         `{"body":"hello"}`,
			Header:       nil,
			WantStatus:   http.StatusNotFound,
			WantResponse: `{"status":404,"error":"Not found","details":"The requested resource does not exist, or you don't have permissions to access it"}`,
		},
		{
			Name:         "creation error",
			Method:       "POST",
			URL:          "/apps/app_id/connections/error/messages",
			Body:         `{"body":"error"}`,
			Header:       nil,
			WantStatus:   http.StatusInternalServerError,
			WantResponse: `There was a problem with your request. *`,
		},
	}
	for _, tc := range tests {
		test.Endpoint(t, router, tc)
	}
}

func TestDeleteMessageAPIEndpointAsPlainWithoutPermissions(t *testing.T) {
	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)

	rg := router.Group("/apps")
	rg.Use(acl.AuthAsPlainMiddleware([]string{}))
	rg.Use(acl.NewMiddleware(filter.NewChecker()).Process)
	RegisterHandlers(rg, mockService{}, mockConnectionService{}, logger)

	tests := []test.APITestCase{
		{
			Name:         "not found",
			Method:       "DELETE",
			URL:          "/apps/app_id/connections/conn_id/messages/id",
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

func TestUpdateMessageAPIEndpointAsPlainWithoutPermissions(t *testing.T) {
	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)

	rg := router.Group("/apps")
	rg.Use(acl.AuthAsPlainMiddleware([]string{}))
	rg.Use(acl.NewMiddleware(filter.NewChecker()).Process)
	RegisterHandlers(rg, mockService{}, mockConnectionService{}, logger)

	tests := []test.APITestCase{
		{
			Name:         "not found",
			Method:       "PUT",
			URL:          "/apps/app_id/connections/conn_id/messages/message_id",
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

func TestUpdateMessageAPIEndpoint(t *testing.T) {
	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)

	rg := router.Group("/apps")
	rg.Use(acl.AuthAsAdminMiddleware())
	rg.Use(acl.NewMiddleware(filter.NewChecker()).Process)
	RegisterHandlers(rg, mockService{}, mockConnectionService{}, logger)

	tests := []test.APITestCase{
		{
			Name:         "success",
			Method:       "PUT",
			URL:          "/apps/app_id/connections/conn_id/messages/message_id",
			Body:         `{"body":"hello"}`,
			Header:       nil,
			WantStatus:   http.StatusOK,
			WantResponse: ``,
		},
		{
			Name:         "invalid input",
			Method:       "PUT",
			URL:          "/apps/app_id/connections/conn_id/messages/message_id",
			Body:         `{]`,
			Header:       nil,
			WantStatus:   http.StatusBadRequest,
			WantResponse: `{"details":"The provided body is not valid", "error":"Invalid input", "status":400}`,
		},
		{
			Name:         "validation error",
			Method:       "PUT",
			URL:          "/apps/app_id/connections/conn_id/messages/message_id",
			Body:         `{}`,
			Header:       nil,
			WantStatus:   http.StatusBadRequest,
			WantResponse: `{"details":"body: cannot be blank.", "error":"Invalid input", "status":400}`,
		},
		{
			Name:         "connection get error",
			Method:       "PUT",
			URL:          "/apps/app_id/connections/not_found_id/messages/message_id",
			Body:         `{"body":"hello"}`,
			Header:       nil,
			WantStatus:   http.StatusNotFound,
			WantResponse: `{"status":404,"error":"Not found","details":"The requested resource does not exist, or you don't have permissions to access it"}`,
		},
		{
			Name:         "modification error",
			Method:       "PUT",
			URL:          "/apps/app_id/connections/conn_id/messages/message_id",
			Body:         `{"body":"error"}`,
			Header:       nil,
			WantStatus:   http.StatusInternalServerError,
			WantResponse: `There was a problem with your request. *`,
		},
	}
	for _, tc := range tests {
		test.Endpoint(t, router, tc)
	}
}
func TestDeleteMessageAPIEndpoint(t *testing.T) {
	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)

	rg := router.Group("/apps")
	rg.Use(acl.AuthAsAdminMiddleware())
	rg.Use(acl.NewMiddleware(filter.NewChecker()).Process)
	RegisterHandlers(rg, mockService{}, mockConnectionService{}, logger)

	tests := []test.APITestCase{
		{
			Name:         "success",
			Method:       "DELETE",
			URL:          "/apps/app_id/connections/conn_id/messages/id",
			Body:         `{"name":"new_name"}`,
			Header:       nil,
			WantStatus:   http.StatusNoContent,
			WantResponse: ``,
		},
		{
			Name:         "modification error",
			Method:       "DELETE",
			URL:          "/apps/app_id/connections/not_found_id/messages/id",
			Body:         `{"name":"controlled_error"}`,
			Header:       nil,
			WantStatus:   http.StatusNotFound,
			WantResponse: `{"status":404,"error":"Not found","details":"The requested resource does not exist, or you don't have permissions to access it"}`,
		},
		{
			Name:         "modification error",
			Method:       "DELETE",
			URL:          "/apps/app_id/connections/conn_id/messages/error",
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
