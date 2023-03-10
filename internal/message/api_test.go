package message

import (
	"net/http"
	"testing"
	"time"

	"github.com/joinself/restful-client/internal/auth"
	"github.com/joinself/restful-client/internal/entity"
	"github.com/joinself/restful-client/internal/test"
	"github.com/joinself/restful-client/pkg/log"
)

func TestAPI(t *testing.T) {
	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)
	repo := &mockRepository{items: []entity.Message{
		{1, "connection", "", "", "", "hello!", time.Now(), time.Now(), time.Now()},
	}}
	authHandler := auth.MockAuthHandler()
	RegisterHandlers(router.Group(""), NewService(repo, logger, nil), authHandler, logger)
	header := auth.MockAuthHeader()

	tests := []test.APITestCase{
		{"get all", "GET", "/connections/connection/messages", "", header, http.StatusOK, `*"total_count":1*`},
		{"get 123", "GET", "/connections/connection/messages/1", "", header, http.StatusOK, `*1*`},
		{"get unknown", "GET", "/connections/connection/messages/1234", "", header, http.StatusNotFound, ""},
		{"create ok", "POST", "/connections/connection/messages", `{"body":"test"}`, header, http.StatusCreated, "*test*"},
		{"create ok count", "GET", "/connections/connection/messages", "", header, http.StatusOK, `*"total_count":2*`},
		{"create auth error", "POST", "/connections/connection/messages", `{"body":"test"}`, nil, http.StatusUnauthorized, ""},
		{"create input error", "POST", "/connections/connection/messages", `"body":"test"}`, header, http.StatusBadRequest, ""},
		{"update ok", "PUT", "/connections/connection/messages/1", `{"body":"messagexyz"}`, header, http.StatusOK, "*messagexyz*"},
		{"update verify", "GET", "/connections/connection/messages/1", "", header, http.StatusOK, `*messagexyz*`},
		{"update auth error", "PUT", "/connections/connection/messages/1", `{"body":"messagexyz"}`, nil, http.StatusUnauthorized, ""},
		{"update input error", "PUT", "/connections/connection/messages/1", `"body":"messagexyz"}`, header, http.StatusBadRequest, ""},
		{"delete ok", "DELETE", "/connections/connection/messages/1", ``, header, http.StatusOK, "*messagexyz*"},
		{"delete verify", "DELETE", "/connections/connection/messages/1", ``, header, http.StatusNotFound, ""},
		{"delete auth error", "DELETE", "/connections/connection/messages/1", ``, nil, http.StatusUnauthorized, ""},
	}
	for _, tc := range tests {
		test.Endpoint(t, router, tc)
	}
}
