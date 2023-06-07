package fact

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
	connRepo := &mock.ConnectionRepositoryMock{Items: []entity.Connection{
		{123, "connection", "app1", "connection123", time.Now(), time.Now()},
	}}

	repo := &mock.FactRepositoryMock{Items: []entity.Fact{
		{"123", 123, "", "cid", "jti", "status", "source", "field", "value", time.Now(), time.Now(), time.Now()},
	}}
	atRepo := &mock.AttestationRepositoryMock{Items: []entity.Attestation{
		{"123", "123", "body", "value", time.Now(), time.Now()},
	}}
	authHandler := auth.MockAuthHandler()
	RegisterHandlers(
		router.Group(""),
		NewService(repo, atRepo, logger, nil),
		connection.NewService(connRepo, logger, nil),
		authHandler,
		logger)
	header := auth.MockAuthHeader()

	tests := []test.APITestCase{
		{"get all", "GET", "/apps/app1/connections/connection/facts", "", header, http.StatusOK, `*"total_count":1*`},
		{"get 123", "GET", "/apps/app1/connections/connection/facts/123", "", header, http.StatusOK, `*123*`},
		{"get unknown", "GET", "/apps/app1/connections/connection/facts/1234", "", header, http.StatusNotFound, ""},
		{"create ok", "POST", "/apps/app1/connections/connection/facts", `{"fact":"test"}`, header, http.StatusCreated, "*test*"},
		{"create ok count", "GET", "/apps/app1/connections/connection/facts", "", header, http.StatusOK, `*"total_count":2*`},
		{"create auth error", "POST", "/apps/app1/connections/connection/facts", `{"body":"test"}`, nil, http.StatusUnauthorized, ""},
		{"create input error", "POST", "/apps/app1/connections/connection/facts", `"body":"test"}`, header, http.StatusBadRequest, ""},
		{"update ok", "PUT", "/apps/app1/connections/connection/facts/123", `{"body":"factxyz"}`, header, http.StatusOK, "*factxyz*"},
		{"update verify", "GET", "/apps/app1/connections/connection/facts/123", "", header, http.StatusOK, `*factxyz*`},
		{"update auth error", "PUT", "/apps/app1/connections/connection/facts/123", `{"body":"factxyz"}`, nil, http.StatusUnauthorized, ""},
		{"update input error", "PUT", "/apps/app1/connections/connection/facts/123", `"body":"factxyz"}`, header, http.StatusBadRequest, ""},
		{"delete ok", "DELETE", "/apps/app1/connections/connection/facts/123", ``, header, http.StatusOK, "*factxyz*"},
		{"delete verify", "DELETE", "/apps/app1/connections/connection/facts/123", ``, header, http.StatusNotFound, ""},
		{"delete auth error", "DELETE", "/apps/app1/connections/connection/facts/123", ``, nil, http.StatusUnauthorized, ""},
	}
	for _, tc := range tests {
		test.Endpoint(t, router, tc)
	}
}
