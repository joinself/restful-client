package signature

import (
	"net/http"
	"testing"

	"github.com/joinself/restful-client/internal/test"
	"github.com/joinself/restful-client/pkg/acl"
	"github.com/joinself/restful-client/pkg/filter"
	"github.com/joinself/restful-client/pkg/log"
)

func TestGetSignatureAPIEndpointAsPlainWithPermissions(t *testing.T) {
	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)

	rg := router.Group("/apps")
	rg.Use(acl.AuthAsPlainMiddleware([]string{"GET /apps/app_id/*"}))
	rg.Use(acl.NewMiddleware(filter.NewChecker()).TokenAndAccessCheckMiddleware)
	RegisterHandlers(rg, mockService{}, mockConnectionService{}, logger)

	tests := []test.APITestCase{
		{
			Name:         "success",
			Method:       "GET",
			URL:          "/apps/app_id/connections/conn_id/signatures/signature_jti",
			Body:         ``,
			Header:       nil,
			WantStatus:   http.StatusOK,
			WantResponse: `{"created_at":"0001-01-01T00:00:00Z", "data":null, "description":"body", "id":"", "signature":"", "status":"", "updated_at":"0001-01-01T00:00:00Z"}`,
		},
		{
			Name:         "connection not found",
			Method:       "GET",
			URL:          "/apps/app_id/connections/not_found_id/signatures/signature_jti",
			Body:         ``,
			Header:       nil,
			WantStatus:   http.StatusNotFound,
			WantResponse: `{"status":404,"error":"Not found","details":"The requested resource does not exist, or you don't have permissions to access it"}`,
		},
		{
			Name:         "connection not found",
			Method:       "GET",
			URL:          "/apps/app_id/connections/conn_id/signatures/not_found_id",
			Body:         ``,
			Header:       nil,
			WantStatus:   http.StatusNotFound,
			WantResponse: `{"status":404,"error":"Not found","details":"The requested resource does not exist, or you don't have permissions to access it"}`,
		},
	}
	for _, tc := range tests {
		test.Endpoint(t, router, tc)
	}
}

func TestGetSignatureAPIEndpointAsPlainWithoutPermissions(t *testing.T) {
	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)

	rg := router.Group("/apps")
	rg.Use(acl.AuthAsPlainMiddleware([]string{}))
	rg.Use(acl.NewMiddleware(filter.NewChecker()).TokenAndAccessCheckMiddleware)
	RegisterHandlers(rg, mockService{}, mockConnectionService{}, logger)

	tests := []test.APITestCase{
		{
			Name:         "success",
			Method:       "GET",
			URL:          "/apps/app_id/connections/conn_id/signatures/signature_jti",
			Body:         ``,
			Header:       nil,
			WantStatus:   http.StatusNotFound,
			WantResponse: `{"status":404,"error":"Not found","details":"The requested resource does not exist, or you don't have permissions to access it"}`,
		},
	}
	for _, tc := range tests {
		test.Endpoint(t, router, tc)
	}
}

func TestListSignaturesAPIEndpointAsAdmin(t *testing.T) {
	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)

	rg := router.Group("/apps")
	rg.Use(acl.AuthAsAdminMiddleware())
	rg.Use(acl.NewMiddleware(filter.NewChecker()).TokenAndAccessCheckMiddleware)
	RegisterHandlers(rg, mockService{}, mockConnectionService{}, logger)

	tests := []test.APITestCase{
		{
			Name:         "success",
			Method:       "GET",
			URL:          "/apps/app_id/connections/conn_id/signatures",
			Body:         ``,
			Header:       nil,
			WantStatus:   http.StatusOK,
			WantResponse: `{"items":[], "page":1, "page_count":0, "per_page":100, "total_count":0}`,
		},
		{
			Name:         "invalid_connection",
			Method:       "GET",
			URL:          "/apps/app_id/connections/not_found_id/signatures",
			Body:         ``,
			Header:       nil,
			WantStatus:   http.StatusNotFound,
			WantResponse: `{"status":404,"error":"Not found","details":"The requested resource does not exist, or you don't have permissions to access it"}`,
		},
		{
			Name:         "internal error on count",
			Method:       "GET",
			URL:          "/apps/app_id/connections/conn_id/signatures?signatures_since=99",
			Body:         ``,
			Header:       nil,
			WantStatus:   http.StatusInternalServerError,
			WantResponse: `There was a problem with your request. *`,
		},
		{
			Name:         "internal error on query",
			Method:       "GET",
			URL:          "/apps/app_id/connections/conn_id/signatures?signatures_since=98",
			Body:         ``,
			Header:       nil,
			WantStatus:   http.StatusInternalServerError,
			WantResponse: `There was a problem with your request. *`,
		},
	}
	for _, tc := range tests {
		test.Endpoint(t, router, tc)
	}
}

func TestListSignaturesAPIEndpointAsPlain(t *testing.T) {
	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)

	rg := router.Group("/apps")
	rg.Use(acl.AuthAsPlainMiddleware([]string{}))
	rg.Use(acl.NewMiddleware(filter.NewChecker()).TokenAndAccessCheckMiddleware)
	RegisterHandlers(rg, mockService{}, mockConnectionService{}, logger)

	tests := []test.APITestCase{
		{
			Name:         "not found",
			Method:       "GET",
			URL:          "/apps/app_id/connections/conn_id/signatures",
			Body:         ``,
			Header:       nil,
			WantStatus:   http.StatusNotFound,
			WantResponse: `{"status":404,"error":"Not found","details":"The requested resource does not exist, or you don't have permissions to access it"}`,
		},
	}
	for _, tc := range tests {
		test.Endpoint(t, router, tc)
	}
}

func TestCreateSignatureAPIEndpointAsPlainWithoutPermissions(t *testing.T) {
	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)

	rg := router.Group("/apps")
	rg.Use(acl.AuthAsPlainMiddleware([]string{}))
	rg.Use(acl.NewMiddleware(filter.NewChecker()).TokenAndAccessCheckMiddleware)
	RegisterHandlers(rg, mockService{}, mockConnectionService{}, logger)

	tests := []test.APITestCase{
		{
			Name:         "not found",
			Method:       "POST",
			URL:          "/apps/app_id/connections/conn_id/signatures",
			Body:         ``,
			Header:       nil,
			WantStatus:   http.StatusNotFound,
			WantResponse: `{"status":404,"error":"Not found","details":"The requested resource does not exist, or you don't have permissions to access it"}`,
		},
	}
	for _, tc := range tests {
		test.Endpoint(t, router, tc)
	}
}

func TestCreateSignatureAPIEndpoint(t *testing.T) {
	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)

	rg := router.Group("/apps")
	rg.Use(acl.AuthAsAdminMiddleware())
	rg.Use(acl.NewMiddleware(filter.NewChecker()).TokenAndAccessCheckMiddleware)
	RegisterHandlers(rg, mockService{}, mockConnectionService{}, logger)

	tests := []test.APITestCase{
		{
			Name:         "success",
			Method:       "POST",
			URL:          "/apps/app_id/connections/conn_id/signatures",
			Body:         `{"description":"hello"}`,
			Header:       nil,
			WantStatus:   http.StatusOK,
			WantResponse: ``,
		},
		{
			Name:         "invalid input",
			Method:       "POST",
			URL:          "/apps/app_id/connections/conn_id/signatures",
			Body:         `{]`,
			Header:       nil,
			WantStatus:   http.StatusBadRequest,
			WantResponse: `{"details":"The provided body is not valid", "error":"Invalid input", "status":400}`,
		},
		{
			Name:         "validation error",
			Method:       "POST",
			URL:          "/apps/app_id/connections/conn_id/signatures",
			Body:         `{}`,
			Header:       nil,
			WantStatus:   http.StatusBadRequest,
			WantResponse: `{"details":"Description: cannot be blank.", "error":"Invalid input", "status":400}`,
		},
		{
			Name:         "connection get error",
			Method:       "POST",
			URL:          "/apps/app_id/connections/not_found_id/signatures",
			Body:         `{"description":"hello"}`,
			Header:       nil,
			WantStatus:   http.StatusNotFound,
			WantResponse: `{"status":404,"error":"Not found","details":"The requested resource does not exist, or you don't have permissions to access it"}`,
		},
		{
			Name:         "success-with-objects",
			Method:       "POST",
			URL:          "/apps/app_id/connections/conn_id/signatures",
			Body:         `{"description":"hello","objects":[{"title":"yolo","data_uri":"data:text/plain;charset=utf-8;base64,SGVsbG8gd29ybGQ="}]}`,
			Header:       nil,
			WantStatus:   http.StatusOK,
			WantResponse: ``,
		},
		{
			Name:         "object-title-validation-error",
			Method:       "POST",
			URL:          "/apps/app_id/connections/conn_id/signatures",
			Body:         `{"description":"hello","objects":[{"title":"y","data_uri":"data:text/plain;charset=utf-8;base64,SGVsbG8gd29ybGQ="}]}`,
			Header:       nil,
			WantStatus:   http.StatusBadRequest,
			WantResponse: `{"details":"title: the length must be between 3 and 128.", "error":"Invalid input", "status":400}`,
		},
		{
			Name:         "object-data-uri-validation-error",
			Method:       "POST",
			URL:          "/apps/app_id/connections/conn_id/signatures",
			Body:         `{"description":"hello","objects":[{"title":"yolo","data_uri":"randomtext"}]}`,
			Header:       nil,
			WantStatus:   http.StatusBadRequest,
			WantResponse: `{"details":"data_uri: must be in a valid format.", "error":"Invalid input", "status":400}`,
		},
	}
	for _, tc := range tests {
		test.Endpoint(t, router, tc)
	}
}

func TestDeleteSignatureAPIEndpointAsPlainWithoutPermissions(t *testing.T) {
	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)

	rg := router.Group("/apps")
	rg.Use(acl.AuthAsPlainMiddleware([]string{}))
	rg.Use(acl.NewMiddleware(filter.NewChecker()).TokenAndAccessCheckMiddleware)
	RegisterHandlers(rg, mockService{}, mockConnectionService{}, logger)

	tests := []test.APITestCase{
		{
			Name:         "not found",
			Method:       "DELETE",
			URL:          "/apps/app_id/connections/conn_id/signatures/id",
			Body:         ``,
			Header:       nil,
			WantStatus:   http.StatusNotFound,
			WantResponse: `{"status":404,"error":"Not found","details":"The requested resource does not exist, or you don't have permissions to access it"}`,
		},
	}
	for _, tc := range tests {
		test.Endpoint(t, router, tc)
	}
}

func TestUpdateSignatureAPIEndpointAsPlainWithoutPermissions(t *testing.T) {
	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)

	rg := router.Group("/apps")
	rg.Use(acl.AuthAsPlainMiddleware([]string{}))
	rg.Use(acl.NewMiddleware(filter.NewChecker()).TokenAndAccessCheckMiddleware)
	RegisterHandlers(rg, mockService{}, mockConnectionService{}, logger)

	tests := []test.APITestCase{
		{
			Name:         "not found",
			Method:       "PUT",
			URL:          "/apps/app_id/connections/conn_id/signatures/signature_id",
			Body:         ``,
			Header:       nil,
			WantStatus:   http.StatusNotFound,
			WantResponse: `{"status":404,"error":"Not found","details":"The requested resource does not exist, or you don't have permissions to access it"}`,
		},
	}
	for _, tc := range tests {
		test.Endpoint(t, router, tc)
	}
}
