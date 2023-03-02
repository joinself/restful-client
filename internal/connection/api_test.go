package connection

import (
	"net/http"
	"testing"
	"time"

	"github.com/qiangxue/go-rest-api/internal/auth"
	"github.com/qiangxue/go-rest-api/internal/entity"
	"github.com/qiangxue/go-rest-api/internal/test"
	"github.com/qiangxue/go-rest-api/pkg/log"
)

func TestAPI(t *testing.T) {
	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)
	repo := &mockRepository{items: []entity.Connection{
		{"123", "connection123", "1112223334", time.Now(), time.Now()},
	}}
	RegisterHandlers(router.Group(""), NewService(repo, logger), auth.MockAuthHandler, logger)
	header := auth.MockAuthHeader()

	tests := []test.APITestCase{
		{"get all", "GET", "/connections", "", header, http.StatusOK, `*"total_count":1*`},
		{"get 123", "GET", "/connections/123", "", header, http.StatusOK, `*connection123*`},
		{"get unknown", "GET", "/connections/1234", "", header, http.StatusNotFound, ""},
		{"create ok", "POST", "/connections", `{"name":"test"}`, header, http.StatusCreated, "*test*"},
		{"create ok count", "GET", "/connections", "", header, http.StatusOK, `*"total_count":2*`},
		{"create auth error", "POST", "/connections", `{"name":"test"}`, nil, http.StatusUnauthorized, ""},
		{"create input error", "POST", "/connections", `"name":"test"}`, header, http.StatusBadRequest, ""},
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
