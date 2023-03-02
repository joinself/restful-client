package fact

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
	repo := &mockRepository{items: []entity.Fact{
		{"123", "connection", "", "source", "value", time.Now(), time.Now(), time.Now()},
	}}
	RegisterHandlers(router.Group(""), NewService(repo, logger), auth.MockAuthHandler, logger)
	header := auth.MockAuthHeader()

	tests := []test.APITestCase{
		{"get all", "GET", "/connections/connection/facts", "", header, http.StatusOK, `*"total_count":1*`},
		{"get 123", "GET", "/connections/connection/facts/123", "", header, http.StatusOK, `*123*`},
		{"get unknown", "GET", "/connections/connection/facts/1234", "", header, http.StatusNotFound, ""},
		{"create ok", "POST", "/connections/connection/facts", `{"body":"test"}`, header, http.StatusCreated, "*test*"},
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
