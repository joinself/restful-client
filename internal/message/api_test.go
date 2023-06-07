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
		{1, 123, "", "", "", "hello!", time.Now(), time.Now(), time.Now()},
	}}
	connRepo := &mock.ConnectionRepositoryMock{Items: []entity.Connection{
		{123, "connection", "app1", "connection123", time.Now(), time.Now()},
	}}
	authHandler := auth.MockAuthHandler()
	RegisterHandlers(
		router.Group(""),
		NewService(repo, logger, nil),
		connection.NewService(connRepo, logger, nil),
		authHandler,
		logger)
	header := auth.MockAuthHeader()

	tests := []test.APITestCase{
		{"get all", "GET", "/apps/app1/connections/connection/messages", "", header, http.StatusOK, `*"total_count":1*`},
		{"get 123", "GET", "/apps/app1/connections/connection/messages/1", "", header, http.StatusOK, `*1*`},
		{"get unknown", "GET", "/apps/app1/connections/connection/messages/1234", "", header, http.StatusNotFound, ""},
		{"create ok", "POST", "/apps/app1/connections/connection/messages", `{"body":"test"}`, header, http.StatusCreated, "*test*"},
		{"create ok count", "GET", "/apps/app1/connections/connection/messages", "", header, http.StatusOK, `*"total_count":2*`},
		{"create auth error", "POST", "/apps/app1/connections/connection/messages", `{"body":"test"}`, nil, http.StatusUnauthorized, ""},
		{"create input error", "POST", "/apps/app1/connections/connection/messages", `"body":"test"}`, header, http.StatusBadRequest, ""},
		{"update ok", "PUT", "/apps/app1/connections/connection/messages/1", `{"body":"messagexyz"}`, header, http.StatusOK, "*messagexyz*"},
		{"update verify", "GET", "/apps/app1/connections/connection/messages/1", "", header, http.StatusOK, `*messagexyz*`},
		{"update auth error", "PUT", "/apps/app1/connections/connection/messages/1", `{"body":"messagexyz"}`, nil, http.StatusUnauthorized, ""},
		{"update input error", "PUT", "/apps/app1/connections/connection/messages/1", `"body":"messagexyz"}`, header, http.StatusBadRequest, ""},
		{"delete ok", "DELETE", "/apps/app1/connections/connection/messages/1", ``, header, http.StatusOK, "*messagexyz*"},
		{"delete verify", "DELETE", "/apps/app1/connections/connection/messages/1", ``, header, http.StatusNotFound, ""},
		{"delete auth error", "DELETE", "/apps/app1/connections/connection/messages/1", ``, nil, http.StatusUnauthorized, ""},
	}
	for _, tc := range tests {
		test.Endpoint(t, router, tc)
	}
}
