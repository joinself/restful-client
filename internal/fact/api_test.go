package fact

import (
	"net/http"
	"testing"
	"time"

	"github.com/joinself/restful-client/internal/connection"
	"github.com/joinself/restful-client/internal/entity"
	"github.com/joinself/restful-client/internal/test"
	"github.com/joinself/restful-client/pkg/acl"
	"github.com/joinself/restful-client/pkg/log"
	"github.com/joinself/restful-client/pkg/mock"
)

func TestAPI(t *testing.T) {
	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)
	connRepo := &mock.ConnectionRepositoryMock{Items: []entity.Connection{
		{ID: 123, SelfID: "connection", AppID: "app1", Name: "connection123", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}}

	repo := &mock.FactRepositoryMock{Items: []entity.Fact{
		{ID: "123", ConnectionID: 123, ISS: "", CID: "cid", JTI: "jti", Status: "status", Source: "source", Fact: "field", Body: "value", IAT: time.Now(), CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}}
	atRepo := &mock.AttestationRepositoryMock{Items: []entity.Attestation{
		{ID: "123", FactID: "123", Body: "body", Value: "value", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}}

	runner := mock.NewRunnerMock()
	RegisterHandlers(
		router.Group("/apps"),
		NewService(repo, atRepo, runner, logger),
		connection.NewService(connRepo, runner, logger),
		logger)
	header := acl.MockAuthHeader()

	tests := []test.APITestCase{
		{Name: "get all", Method: "GET", URL: "/apps/app1/connections/connection/facts", Body: "", Header: header, WantStatus: http.StatusOK, WantResponse: `*"total_count":1*`},
		{Name: "get 123", Method: "GET", URL: "/apps/app1/connections/connection/facts/123", Body: "", Header: header, WantStatus: http.StatusOK, WantResponse: `*123*`},
		{Name: "get unknown", Method: "GET", URL: "/apps/app1/connections/connection/facts/1234", Body: "", Header: header, WantStatus: http.StatusNotFound, WantResponse: ""},
		{Name: "create ok", Method: "POST", URL: "/apps/app1/connections/connection/facts", Body: `{"fact":"test"}`, Header: header, WantStatus: http.StatusCreated, WantResponse: ""},
		{Name: "create input error", Method: "POST", URL: "/apps/app1/connections/connection/facts", Body: `"body":"test"}`, Header: header, WantStatus: http.StatusBadRequest, WantResponse: ""},
	}
	for _, tc := range tests {
		test.Endpoint(t, router, tc)
	}
}
