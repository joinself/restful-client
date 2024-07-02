package connection

import (
	"net/http"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/joinself/restful-client/pkg/response"
)

// ExtListResponse represents the json object returned when listing connections.
type ExtListResponse struct {
	Page       int             `json:"page"`
	PerPage    int             `json:"per_page"`
	PageCount  int             `json:"page_count"`
	TotalCount int             `json:"total_count"`
	Items      []ExtConnection `json:"items"`
}

// ExtConnection external representation of a connection.
type ExtConnection struct {
	ID        string    `json:"id"`
	AppID     string    `json:"app_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateConnectionRequest represents an connection creation request.
type CreateConnectionRequest struct {
	SelfID string `json:"selfid"`
}

// Validate validates the CreateConnectionRequest fields.
func (m CreateConnectionRequest) Validate() *response.Error {
	err := validation.ValidateStruct(&m,
		validation.Field(&m.SelfID, validation.Required, validation.Length(3, 128)),
	)
	if err == nil {
		return nil
	}
	return &response.Error{
		Status:  http.StatusBadRequest,
		Error:   "Invalid input",
		Details: err.Error(),
	}
}

// UpdateConnectionRequest represents an connection update request.
type UpdateConnectionRequest struct {
	Name string `json:"name"`
}

// Validate validates the CreateConnectionRequest fields.
func (m UpdateConnectionRequest) Validate() *response.Error {
	err := validation.ValidateStruct(&m,
		validation.Field(&m.Name, validation.Required, validation.Length(3, 128)),
	)
	if err == nil {
		return nil
	}
	return &response.Error{
		Status:  http.StatusBadRequest,
		Error:   "Invalid input",
		Details: err.Error(),
	}
}
