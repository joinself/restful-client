package group

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
	room := entity.Room{
		ID:        123,
		Appid:     "app1",
		Name:      "group123",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	repo := &mock.GroupRepositoryMock{Items: []mock.Item{
		{
			Room:    &room,
			Members: []string{},
		},
	}}

	cRepo := &mock.ConnectionRepositoryMock{Items: []entity.Connection{{
		ID:        1112223334,
		SelfID:    "1112223334",
		AppID:     "app1",
		Name:      "connection1112223334",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, {
		ID:        1112223335,
		SelfID:    "1112223335",
		AppID:     "app1",
		Name:      "connection1112223335",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}}}

	authHandler := auth.MockAuthHandler()
	s := NewService(repo, cRepo, logger, nil)
	RegisterHandlers(router.Group(""), s, authHandler, logger)
	header := auth.MockAuthHeader()

	tests := []test.APITestCase{
		{
			Name:         "get all",
			Method:       "GET",
			URL:          "/apps/app1/groups",
			Body:         "",
			Header:       header,
			WantStatus:   http.StatusOK,
			WantResponse: `*"total_count":1*`,
		}, {
			Name:         "get 123",
			Method:       "GET",
			URL:          "/apps/app1/groups/123",
			Body:         "",
			Header:       header,
			WantStatus:   http.StatusOK,
			WantResponse: `*group123*`,
		}, {
			Name:         "get unknown",
			Method:       "GET",
			URL:          "/apps/app1/groups/999",
			Body:         "",
			Header:       header,
			WantStatus:   http.StatusNotFound,
			WantResponse: "",
		}, {
			Name:         "create ok",
			Method:       "POST",
			URL:          "/apps/app1/groups",
			Body:         `{"name": "sid1", "members": ["1112223334", "1112223335"]}`,
			Header:       header,
			WantStatus:   http.StatusCreated,
			WantResponse: "*sid1*",
		}, {
			Name:         "create ok count",
			Method:       "GET",
			URL:          "/apps/app1/groups",
			Body:         "",
			Header:       header,
			WantStatus:   http.StatusOK,
			WantResponse: `*"total_count":2*`,
		}, {
			Name:         "create auth error",
			Method:       "POST",
			URL:          "/apps/app1/groups",
			Body:         `{"name":"broken", "members": ["1112223334", "1112223335"]}`,
			Header:       nil,
			WantStatus:   http.StatusUnauthorized,
			WantResponse: "",
		}, {
			Name:         "create input error",
			Method:       "POST",
			URL:          "/apps/app1/groups",
			Body:         `"name":"test"}`,
			Header:       header,
			WantStatus:   http.StatusBadRequest,
			WantResponse: "",
		}, {
			Name:         "update ok",
			Method:       "PUT",
			URL:          "/apps/app1/groups/123",
			Body:         `{"name":"groupxyz", "members": ["1112223334", "1112223335"]}`,
			Header:       header,
			WantStatus:   http.StatusOK,
			WantResponse: "*groupxyz*",
		}, {
			Name:         "update verify",
			Method:       "GET",
			URL:          "/apps/app1/groups/123",
			Body:         "",
			Header:       header,
			WantStatus:   http.StatusOK,
			WantResponse: `*groupxyz*`,
		}, {
			Name:         "update auth error",
			Method:       "PUT",
			URL:          "/apps/app1/groups/123",
			Body:         `{"name":"groupxyz"}`,
			Header:       nil,
			WantStatus:   http.StatusUnauthorized,
			WantResponse: "",
		}, {
			Name:         "update input error",
			Method:       "PUT",
			URL:          "/apps/app1/groups/123",
			Body:         `"name":"groupxyz"}`,
			Header:       header,
			WantStatus:   http.StatusBadRequest,
			WantResponse: "",
		}, {
			Name:         "delete ok",
			Method:       "DELETE",
			URL:          "/apps/app1/groups/123",
			Body:         ``,
			Header:       header,
			WantStatus:   http.StatusOK,
			WantResponse: "",
		}, {
			Name:         "delete verify",
			Method:       "DELETE",
			URL:          "/apps/app1/groups/123",
			Body:         ``,
			Header:       header,
			WantStatus:   http.StatusNotFound,
			WantResponse: "",
		}, {
			Name:         "delete auth error",
			Method:       "DELETE",
			URL:          "/apps/app1/groups/123",
			Body:         ``,
			Header:       nil,
			WantStatus:   http.StatusUnauthorized,
			WantResponse: "",
		},
	}
	for _, tc := range tests {
		test.Endpoint(t, router, tc)
	}
}
