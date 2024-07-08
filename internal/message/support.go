package message

import (
	"net/http"
	"regexp"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/joinself/restful-client/pkg/response"
)

type ExtListResponse struct {
	Page       int       `json:"page"`
	PerPage    int       `json:"per_page"`
	PageCount  int       `json:"page_count"`
	TotalCount int       `json:"total_count"`
	Items      []Message `json:"items"`
}

type MessageObject struct {
	Link    string `json:"link"`
	Name    string `json:"name"`
	Mime    string `json:"mime"`
	Expires int64  `json:"expires,omitempty"`
	Key     string `json:"key,omitempty"`
}

func (m MessageObject) Validate() error {
	mimePattern := `^[\w\-\+\.]+\/[\w\-\+\.]+$`

	return validation.ValidateStruct(&m,
		validation.Field(&m.Link, validation.Required, is.URL),
		validation.Field(&m.Name, validation.Required, validation.Length(3, 128)),
		validation.Field(&m.Mime, validation.Required, validation.Match(regexp.MustCompile(mimePattern))),
		validation.Field(&m.Key, validation.Required, validation.Length(10, 128)),
	)
}

type MessageOptions struct {
	Objects []MessageObject `json:"objects"`
}

type CreateMessageRequest struct {
	Body    string         `json:"body"`
	Options MessageOptions `json:"options,omitempty"`
}

// Validate validates the CreateMessageRequest fields.
func (m CreateMessageRequest) Validate() *response.Error {
	err := validation.ValidateStruct(&m,
		validation.Field(&m.Body, validation.Required, validation.Length(0, 128)),
	)
	if err != nil {
		return &response.Error{
			Status:  http.StatusBadRequest,
			Error:   "Invalid input",
			Details: err.Error(),
		}
	}

	err = validateObjects(m.Options.Objects)
	if err != nil {
		return &response.Error{
			Status:  http.StatusBadRequest,
			Error:   "Invalid input",
			Details: err.Error(),
		}
	}

	return nil
}

// UpdateMessageRequest represents an message update request.
type UpdateMessageRequest struct {
	Body string `json:"body"`
}

// Validate validates the CreateMessageRequest fields.
func (m UpdateMessageRequest) Validate() *response.Error {
	err := validation.ValidateStruct(&m,
		validation.Field(&m.Body, validation.Required, validation.Length(0, 128)),
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

func validateObjects(objects []MessageObject) error {
	for _, o := range objects {
		if err := o.Validate(); err != nil {
			return err
		}
	}
	return nil
}
