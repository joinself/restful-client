package acl

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/joinself/restful-client/pkg/filter"
	"github.com/joinself/restful-client/pkg/response"
	"github.com/labstack/echo/v4"
)

type (
	Middleware struct {
		checker *filter.Checker
		mutex   sync.RWMutex
	}
)

func NewMiddleware(checker *filter.Checker) *Middleware {
	return &Middleware{
		checker: checker,
	}
}

// TokenAndAccessCheckMiddleware is the middleware function.
func (s *Middleware) TokenAndAccessCheckMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		s.mutex.Lock()
		defer s.mutex.Unlock()

		tok, ok := CurrentToken(c)
		if !ok {
			return c.JSON(http.StatusNotFound, response.Error{
				Status:  http.StatusNotFound,
				Error:   "Not found",
				Details: "The requested resource does not exist, or you don't have permissions to access it",
			})
		}
		if s.checker.Check(tok) { // if it's blacklisted...
			return c.JSON(http.StatusNotFound, response.Error{
				Status:  http.StatusNotFound,
				Error:   "Not found",
				Details: "The requested resource does not exist, or you don't have permissions to access it",
			})
		}

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

		fullResource := fmt.Sprintf("%s %s", c.Request().Method, c.Request().URL.String())
		if !HasAccessToResource(c, fullResource) {
			return nil
		}

		return next(c)
	}
}
