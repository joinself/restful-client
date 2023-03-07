package fact

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
	repo := &mockRepository{items: []entity.Fact{
		{"123", "connection", "", "cid", "jti", "status", "source", "field", "value", time.Now(), time.Now(), time.Now()},
	}}
	atRepo := &mockAtRepository{items: []entity.Attestation{
		{"123", "123", "body", "value", time.Now(), time.Now()},
	}}
	RegisterHandlers(router.Group(""), NewService(repo, atRepo, logger, nil), auth.MockAuthHandler, logger)
	header := auth.MockAuthHeader()

	tests := []test.APITestCase{
		{"get all", "GET", "/connections/connection/facts", "", header, http.StatusOK, `*"total_count":1*`},
		{"get 123", "GET", "/connections/connection/facts/123", "", header, http.StatusOK, `*123*`},
		{"get unknown", "GET", "/connections/connection/facts/1234", "", header, http.StatusNotFound, ""},
		{"create ok", "POST", "/connections/connection/facts", `{"fact":"test"}`, header, http.StatusCreated, "*test*"},
		{"create ok count", "GET", "/connections/connection/facts", "", header, http.StatusOK, `*"total_count":2*`},
		{"create auth error", "POST", "/connections/connection/facts", `{"body":"test"}`, nil, http.StatusUnauthorized, ""},
		{"create input error", "POST", "/connections/connection/facts", `"body":"test"}`, header, http.StatusBadRequest, ""},
		{"update ok", "PUT", "/connections/connection/facts/123", `{"body":"factxyz"}`, header, http.StatusOK, "*factxyz*"},
		{"update verify", "GET", "/connections/connection/facts/123", "", header, http.StatusOK, `*factxyz*`},
		{"update auth error", "PUT", "/connections/connection/facts/123", `{"body":"factxyz"}`, nil, http.StatusUnauthorized, ""},
		{"update input error", "PUT", "/connections/connection/facts/123", `"body":"factxyz"}`, header, http.StatusBadRequest, ""},
		{"delete ok", "DELETE", "/connections/connection/facts/123", ``, header, http.StatusOK, "*factxyz*"},
		{"delete verify", "DELETE", "/connections/connection/facts/123", ``, header, http.StatusNotFound, ""},
		{"delete auth error", "DELETE", "/connections/connection/facts/123", ``, nil, http.StatusUnauthorized, ""},
	}
	for _, tc := range tests {
		test.Endpoint(t, router, tc)
	}
}
