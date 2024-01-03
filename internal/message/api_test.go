package message

import (
	"net/http"
	"testing"
	"time"

	"github.com/joinself/restful-client/internal/auth"
	"github.com/joinself/restful-client/internal/connection"
	"github.com/joinself/restful-client/internal/entity"
	"github.com/joinself/restful-client/internal/test"
	"github.com/joinself/restful-client/pkg/log"
	"github.com/joinself/restful-client/pkg/mock"
)

func TestAPI(t *testing.T) {
	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)
	repo := &mock.MessageRepositoryMock{Items: []entity.Message{
		{ID: 1, JTI: "1", ConnectionID: 123, ISS: "", CID: "", RID: "", Body: "hello!", IAT: time.Now(), CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}}
	connRepo := &mock.ConnectionRepositoryMock{Items: []entity.Connection{
		{ID: 123, SelfID: "connection", AppID: "app1", Name: "connection123", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}}
	authHandler := auth.MockAuthHandler()
	runner := mock.NewRunnerMock()
	RegisterHandlers(
		router.Group(""),
		NewService(repo, runner, logger),
		connection.NewService(connRepo, runner, logger),
		authHandler,
		logger)
	header := auth.MockAuthHeader()

	tests := []test.APITestCase{
		{Name: "get all", Method: "GET", URL: "/apps/app1/connections/connection/messages", Body: "", Header: header, WantStatus: http.StatusOK, WantResponse: `*"total_count":1*`},
		{Name: "get 123", Method: "GET", URL: "/apps/app1/connections/connection/messages/1", Body: "", Header: header, WantStatus: http.StatusOK, WantResponse: `*1*`},
		{Name: "get unknown", Method: "GET", URL: "/apps/app1/connections/connection/messages/1234", Body: "", Header: header, WantStatus: http.StatusNotFound, WantResponse: ""},
		{Name: "create ok", Method: "POST", URL: "/apps/app1/connections/connection/messages", Body: `{"body":"test"}`, Header: header, WantStatus: http.StatusCreated, WantResponse: "*test*"},
		{Name: "create ok count", Method: "GET", URL: "/apps/app1/connections/connection/messages", Body: "", Header: header, WantStatus: http.StatusOK, WantResponse: `*"total_count":2*`},
		{Name: "create auth error", Method: "POST", URL: "/apps/app1/connections/connection/messages", Body: `{"body":"test"}`, Header: nil, WantStatus: http.StatusUnauthorized, WantResponse: ""},
		{Name: "create input error", Method: "POST", URL: "/apps/app1/connections/connection/messages", Body: `"body":"test"}`, Header: header, WantStatus: http.StatusBadRequest, WantResponse: ""},
		{Name: "update ok", Method: "PUT", URL: "/apps/app1/connections/connection/messages/1", Body: `{"body":"messagexyz"}`, Header: header, WantStatus: http.StatusOK, WantResponse: "*messagexyz*"},
		{Name: "update verify", Method: "GET", URL: "/apps/app1/connections/connection/messages/1", Body: "", Header: header, WantStatus: http.StatusOK, WantResponse: `*messagexyz*`},
		{Name: "update auth error", Method: "PUT", URL: "/apps/app1/connections/connection/messages/1", Body: `{"body":"messagexyz"}`, Header: nil, WantStatus: http.StatusUnauthorized, WantResponse: ""},
		{Name: "update input error", Method: "PUT", URL: "/apps/app1/connections/connection/messages/1", Body: `"body":"messagexyz"}`, Header: header, WantStatus: http.StatusBadRequest, WantResponse: ""},
		{Name: "delete ok", Method: "DELETE", URL: "/apps/app1/connections/connection/messages/1", Body: ``, Header: header, WantStatus: http.StatusOK, WantResponse: ""},
		{Name: "delete verify", Method: "DELETE", URL: "/apps/app1/connections/connection/messages/1", Body: ``, Header: header, WantStatus: http.StatusNotFound, WantResponse: ""},
		{Name: "delete auth error", Method: "DELETE", URL: "/apps/app1/connections/connection/messages/1", Body: ``, Header: nil, WantStatus: http.StatusUnauthorized, WantResponse: ""},
	}
	for _, tc := range tests {
		test.Endpoint(t, router, tc)
	}
}
