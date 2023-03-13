package healthcheck

import (
	"net/http"
	"testing"

	"github.com/joinself/restful-client/internal/test"
	"github.com/joinself/restful-client/pkg/log"
)

func TestAPI(t *testing.T) {
	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)
	RegisterHandlers(router.Group(""), "0.9.0")
	test.Endpoint(t, router, test.APITestCase{
		"ok", "GET", "/healthcheck", "", nil, http.StatusOK, `"OK"`,
	})
}
