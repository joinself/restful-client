package connection

import (
	"net/http"
	"testing"
	"time"

	"github.com/joinself/restful-client/internal/auth"
	"github.com/joinself/restful-client/internal/entity"
	"github.com/joinself/restful-client/internal/test"
	"github.com/joinself/restful-client/pkg/log"
	"github.com/joinself/restful-client/pkg/mock"
)

func TestAPI(t *testing.T) {
	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)
	repo := &mock.ConnectionRepositoryMock{Items: []entity.Connection{
		{"123", "connection1", "app1", "connection123", time.Now(), time.Now()},
	}}
	authHandler := auth.MockAuthHandler()
	RegisterHandlers(router.Group(""), NewService(repo, logger, nil), authHandler, logger)
	header := auth.MockAuthHeader()

	tests := []test.APITestCase{
		{"get all", "GET", "/apps/app1/connections", "", header, http.StatusOK, `*"total_count":1*`},
		{"get 123", "GET", "/apps/app1/connections/connection1", "", header, http.StatusOK, `*connection123*`},
		{"get unknown", "GET", "/apps/app1/connections/1234", "", header, http.StatusNotFound, ""},
		{"create ok", "POST", "/apps/app1/connections", `{"selfid": "sid1"}`, header, http.StatusCreated, "*sid1*"},
		{"create ok count", "GET", "/apps/app1/connections", "", header, http.StatusOK, `*"total_count":2*`},
		{"create auth error", "POST", "/apps/app1/connections", `{"selfid":"test"}`, nil, http.StatusUnauthorized, ""},
		{"create input error", "POST", "/apps/app1/connections", `"selfid":"test"}`, header, http.StatusBadRequest, ""},
		{"update ok", "PUT", "/apps/app1/connections/connection1", `{"name":"connectionxyz"}`, header, http.StatusOK, "*connectionxyz*"},
		{"update verify", "GET", "/apps/app1/connections/connection1", "", header, http.StatusOK, `*connectionxyz*`},
		{"update auth error", "PUT", "/apps/app1/connections/connection1", `{"name":"connectionxyz"}`, nil, http.StatusUnauthorized, ""},
		{"update input error", "PUT", "/apps/app1/connections/connection1", `"name":"connectionxyz"}`, header, http.StatusBadRequest, ""},
		{"delete ok", "DELETE", "/apps/app1/connections/connection1", ``, header, http.StatusOK, "*connectionxyz*"},
		{"delete verify", "DELETE", "/apps/app1/connections/connection1", ``, header, http.StatusNotFound, ""},
		{"delete auth error", "DELETE", "/apps/app1/connections/connection1", ``, nil, http.StatusUnauthorized, ""},
	}
	for _, tc := range tests {
		test.Endpoint(t, router, tc)
	}
}
