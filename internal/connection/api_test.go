package connection

import (
	"net/http"
	"testing"
	"time"

	"github.com/joinself/restful-client/internal/entity"
	"github.com/joinself/restful-client/internal/test"
	"github.com/joinself/restful-client/pkg/acl"
	"github.com/joinself/restful-client/pkg/log"
	"github.com/joinself/restful-client/pkg/mock"
)

func TestAPI(t *testing.T) {
	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)
	repo := &mock.ConnectionRepositoryMock{Items: []entity.Connection{{
		ID:        123,
		SelfID:    "connection1",
		AppID:     "app1",
		Name:      "connection123",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now()},
	}}

	runner := mock.NewRunnerMock()
	RegisterHandlers(router.Group("/apps"), NewService(repo, runner, logger), logger)
	header := acl.MockAuthHeader()

	tests := []test.APITestCase{
		{
			Name:         "get all",
			Method:       "GET",
			URL:          "/apps/app1/connections",
			Body:         "",
			Header:       header,
			WantStatus:   http.StatusOK,
			WantResponse: `*"total_count":1*`,
		}, {
			Name:         "get 123",
			Method:       "GET",
			URL:          "/apps/app1/connections/connection1",
			Body:         "",
			Header:       header,
			WantStatus:   http.StatusOK,
			WantResponse: `*connection123*`,
		}, {
			Name:         "get unknown",
			Method:       "GET",
			URL:          "/apps/app1/connections/1234",
			Body:         "",
			Header:       header,
			WantStatus:   http.StatusNotFound,
			WantResponse: "",
		},
		{
			Name:         "create ok",
			Method:       "POST",
			URL:          "/apps/app1/connections",
			Body:         `{"selfid": "sid1"}`,
			Header:       header,
			WantStatus:   http.StatusCreated,
			WantResponse: "*sid1*",
		},
		{
			Name:         "create ok count",
			Method:       "GET",
			URL:          "/apps/app1/connections",
			Body:         "",
			Header:       header,
			WantStatus:   http.StatusOK,
			WantResponse: `*"total_count":2*`,
		},
		{
			Name:         "create input error",
			Method:       "POST",
			URL:          "/apps/app1/connections",
			Body:         `"selfid":"test"}`,
			Header:       header,
			WantStatus:   http.StatusBadRequest,
			WantResponse: "",
		},
		{
			Name:         "update ok",
			Method:       "PUT",
			URL:          "/apps/app1/connections/connection1",
			Body:         `{"name":"connectionxyz"}`,
			Header:       header,
			WantStatus:   http.StatusOK,
			WantResponse: "*connectionxyz*",
		},
		{
			Name:         "update verify",
			Method:       "GET",
			URL:          "/apps/app1/connections/connection1",
			Body:         "",
			Header:       header,
			WantStatus:   http.StatusOK,
			WantResponse: `*connectionxyz*`,
		},
		{
			Name:         "update input error",
			Method:       "PUT",
			URL:          "/apps/app1/connections/connection1",
			Body:         `"name":"connectionxyz"}`,
			Header:       header,
			WantStatus:   http.StatusBadRequest,
			WantResponse: "",
		},
		{
			Name:         "delete ok",
			Method:       "DELETE",
			URL:          "/apps/app1/connections/connection1",
			Body:         ``,
			Header:       header,
			WantStatus:   http.StatusOK,
			WantResponse: "*connectionxyz*",
		},
		{
			Name:         "delete verify",
			Method:       "DELETE",
			URL:          "/apps/app1/connections/connection1",
			Body:         ``,
			Header:       header,
			WantStatus:   http.StatusNotFound,
			WantResponse: "",
		},
	}
	for _, tc := range tests {
		test.Endpoint(t, router, tc)
	}
}
