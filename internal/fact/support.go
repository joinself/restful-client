package fact

import (
	"net/http"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/joinself/restful-client/pkg/response"
)

type ExtFact struct {
	ISS       string    `json:"iss"`
	Key       string    `json:"key"`
	Source    string    `json:"source"`
	CreatedAt time.Time `json:"created_at"`
	Values    []string  `json:"values"`
	// TODO: is this something that the user provides on the response?
	// Group     string `json:"group"`
}

func NewExtFact(f Fact) ExtFact {
	output := ExtFact{
		ISS:       f.ISS,
		Key:       f.Fact.Fact,
		Source:    f.Source,
		CreatedAt: f.CreatedAt,
		Values:    []string{},
	}
	for _, a := range f.Attestations {
		output.Values = append(output.Values, a.Value)
	}
	return output
}

type ExtListResponse struct {
	Page       int       `json:"page"`
	PerPage    int       `json:"per_page"`
	PageCount  int       `json:"page_count"`
	TotalCount int       `json:"total_count"`
	Items      []ExtFact `json:"items"`
}

// WARNING: Do not use for code purposes, this is only used to generate
// the documentation for the openapi, which seems to be broken for nested
// structs.
type CreateFactRequestDoc struct {
	Facts []struct {
		Key    string `json:"key"`
		Value  string `json:"value"`
		Source string `json:"source"`
		Group  *struct {
			Name string `json:"name"`
			Icon string `json:"icon"`
		} `json:"group,omitempty"`
		Type string `json:"type,omitempty"`
	} `json:"facts"`
}

// CreateFactRequest represents an fact creation request.
type CreateFactRequest struct {
	Facts []FactToIssue `json:"facts"`
}

// Validate validates the CreateFactRequest fields.
func (m CreateFactRequest) Validate() *response.Error {
	if len(m.Facts) == 0 {
		return &response.Error{
			Status:  http.StatusBadRequest,
			Error:   "Invalid input",
			Details: "You should provide at least a fact to be issued",
		}
	}

	for _, f := range m.Facts {
		if err := validateIssuedFact(f); err != nil {
			return &response.Error{
				Status:  http.StatusBadRequest,
				Error:   "Invalid input",
				Details: err.Error(),
			}
		}
	}
	return nil
}

func validateIssuedFact(f FactToIssue) error {
	err := validation.ValidateStruct(&f,
		validation.Field(&f.Key, validation.Required, validation.Length(3, 128)),
		validation.Field(&f.Value, validation.Required, validation.Length(3, 128)),
		validation.Field(&f.Source, validation.Length(0, 128)),
		validation.Field(&f.Type, validation.Length(0, 128)),
	)
	if err != nil {
		return validation.NewError("validation_fact_to_issue", err.Error())
	}

	if f.Group != nil {
		err := validation.ValidateStruct(f.Group,
			validation.Field(&f.Group.Name, validation.Required, validation.Length(3, 128)),
			validation.Field(&f.Group.Icon, validation.Length(0, 128)),
		)
		if err != nil {
			return validation.NewError("validation_fact_to_issue", err.Error())
		}
	}

	if f.Type != "" {
		err := validation.ValidateStruct(&f,
			validation.Field(&f.Type, validation.In("string", "password", "delegation_certificate")),
		)
		if err != nil {
			return validation.NewError("validation_fact_to_issue", err.Error())
		}
	}

	return nil
}
