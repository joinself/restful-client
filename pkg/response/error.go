package response

import (
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/joinself/restful-client/pkg/log"
	"github.com/labstack/echo/v4"
)

// Error used to respond to errored http requests.
type Error struct {
	Status  int    `json:"status"`
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
	Path    string `json:"path,omitempty"`
}

// DefaultInternalError is a function that logs and returns an internal server error.
// It accepts an echo.Context, a log.Logger, and an error message in string format.
// The function generates a unique error code using uuid.NewV4().
// The error message and error code are logged using the logger instance.
// The function then returns an HTTP status code for internal server error (500)
// and an Error instance with the status set to the HTTP status code,
// the error message set to "Internal error",
// and the details containing a message indicating a problem with the request and the generated error code.
func DefaultInternalError(c echo.Context, logger log.Logger, errorMessage string) (int, Error) {
	errorCode, _ := uuid.NewV4()
	logger.With(c.Request().Context()).Info("[%s] %s", errorCode, errorMessage)

	return http.StatusInternalServerError, Error{
		Status:  http.StatusInternalServerError,
		Error:   "Internal error",
		Details: "There was a problem with your request. Error code [" + errorCode.String() + "]",
	}
}

// DefaultNotFoundError function returns a default 404 Not Found HTTP status code and a custom error.
// It signifies that the requested resource does not exist, or the user does not have permission to access it.
func DefaultNotFoundError() (int, Error) {
	return http.StatusNotFound, Error{
		Status:  http.StatusNotFound,
		Error:   "Not found",
		Details: "The requested resource does not exist, or you don't have permissions to access it",
	}
}

// DefaultBadRequestError function returns a default 400 Bad Request HTTP status code and a custom error.
// It signifies that the provided request body is not valid.
func DefaultBadRequestError() (int, Error) {
	return http.StatusBadRequest, Error{
		Status:  http.StatusBadRequest,
		Error:   "Invalid input",
		Details: "The provided body is not valid",
	}
}
