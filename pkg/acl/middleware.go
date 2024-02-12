package acl

import (
	"net/http"
	"sync"

	"github.com/joinself/restful-client/pkg/response"
	"github.com/labstack/echo/v4"
)

type (
	Middleware struct {
		mutex sync.RWMutex
	}
)

func NewMiddleware() *Middleware {
	return &Middleware{}
}

// Process is the middleware function.
func (s *Middleware) Process(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		s.mutex.Lock()
		defer s.mutex.Unlock()

		r := c.Param("app_id")
		if len(r) == 0 {
			if IsAdmin(c) {
				return next(c)
			} else {
				return c.JSON(http.StatusNotFound, response.Error{
					Status:  http.StatusNotFound,
					Error:   "Not found",
					Details: "The requested resource does not exist, or you don't have permissions to access it",
				})
			}
		}

		if !HasAccessToResource(c, r) {
			return nil
		}

		return next(c)
	}
}
