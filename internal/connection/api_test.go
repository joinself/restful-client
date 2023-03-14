package connection

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
	repo := &mockRepository{items: []entity.Connection{
		{"123", "connection123", time.Now(), time.Now()},
	}}
	authHandler := auth.MockAuthHandler()
	RegisterHandlers(router.Group(""), NewService(repo, logger, nil), authHandler, logger)
	header := auth.MockAuthHeader()

	tests := []test.APITestCase{
		{"get all", "GET", "/connections", "", header, http.StatusOK, `*"total_count":1*`},
		{"get 123", "GET", "/connections/123", "", header, http.StatusOK, `*connection123*`},
		{"get unknown", "GET", "/connections/1234", "", header, http.StatusNotFound, ""},
		{"create ok", "POST", "/connections", `{"selfid": "sid1"}`, header, http.StatusCreated, "*sid1*"},
		{"create ok count", "GET", "/connections", "", header, http.StatusOK, `*"total_count":2*`},
		{"create auth error", "POST", "/connections", `{"selfid":"test"}`, nil, http.StatusUnauthorized, ""},
		{"create input error", "POST", "/connections", `"selfid":"test"}`, header, http.StatusBadRequest, ""},
		{"update ok", "PUT", "/connections/123", `{"name":"connectionxyz"}`, header, http.StatusOK, "*connectionxyz*"},
		{"update verify", "GET", "/connections/123", "", header, http.StatusOK, `*connectionxyz*`},
		{"update auth error", "PUT", "/connections/123", `{"name":"connectionxyz"}`, nil, http.StatusUnauthorized, ""},
		{"update input error", "PUT", "/connections/123", `"name":"connectionxyz"}`, header, http.StatusBadRequest, ""},
		{"delete ok", "DELETE", "/connections/123", ``, header, http.StatusOK, "*connectionxyz*"},
		{"delete verify", "DELETE", "/connections/123", ``, header, http.StatusNotFound, ""},
		{"delete auth error", "DELETE", "/connections/123", ``, nil, http.StatusUnauthorized, ""},
	}
	for _, tc := range tests {
		test.Endpoint(t, router, tc)
	}
}
