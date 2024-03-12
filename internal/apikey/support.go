package apikey

import (
	"fmt"
	"net/http"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/joinself/restful-client/pkg/response"
)

// ExtListResponse represents the json object returned when listing apikeys.
type ExtListResponse struct {
	Page       int         `json:"page"`
	PerPage    int         `json:"per_page"`
	PageCount  int         `json:"page_count"`
	TotalCount int         `json:"total_count"`
	Items      []ExtApiKey `json:"items"`
}

// ExtApiKey external representation of a apikey.
type ExtApiKey struct {
	ID        int       `json:"id"`
	AppID     string    `json:"app_id"`
	Name      string    `json:"name"`
	Token     string    `json:"token"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateApiKeyRequest represents an apikey creation request.
type CreateApiKeyRequest struct {
	Name  string `json:"name"`
	Scope string `json:"scope"`
}

// Validate validates the CreateApiKeyRequest fields.
func (m CreateApiKeyRequest) Validate() *response.Error {
	err := validation.ValidateStruct(&m,
		validation.Field(&m.Name, validation.Required, validation.Length(0, 128)),
		validation.Field(&m.Scope, validation.Required, validation.Length(0, 128)),
		validation.Field(&m.Scope, validation.In("FULL", "MESSAGING", "REQUESTS", "METRICS")),
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

func (m CreateApiKeyRequest) GetResources(appID string) []string {
	validResources := map[string][]string{
		"FULL": []string{
			fmt.Sprintf("ANY /v1/apps/%s*", appID),
		},
		"MESSAGING": []string{
			fmt.Sprintf("GET /v1/apps/%s/connections", appID),
			fmt.Sprintf("ANY /v1/apps/%s/connections/*/messages*", appID),
		},
		"REQUESTS": []string{
			fmt.Sprintf("GET /v1/apps/%s/requests*", appID),
		},
		"METRICS": []string{
			fmt.Sprintf("GET /v1/apps/%s/metrics*", appID),
		},
	}

	return validResources[m.Scope]
}

// UpdateApiKeyRequest represents an apikey update request.
type UpdateApiKeyRequest struct {
	Name string `json:"name"`
}

// Validate validates the CreateApiKeyRequest fields.
func (m UpdateApiKeyRequest) Validate() *response.Error {
	err := validation.ValidateStruct(&m,
		validation.Field(&m.Name, validation.Required, validation.Length(0, 128)),
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
