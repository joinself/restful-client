package fact

import (
	"context"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/joinself/restful-client/internal/entity"
	"github.com/joinself/restful-client/pkg/log"
)

// Service encapsulates usecase logic for facts.
type Service interface {
	Get(ctx context.Context, id string) (Fact, error)
	Query(ctx context.Context, connection string, offset, limit int) ([]Fact, error)
	Count(ctx context.Context) (int, error)
	Create(ctx context.Context, connection string, input CreateFactRequest) (Fact, error)
	Update(ctx context.Context, id string, input UpdateFactRequest) (Fact, error)
	Delete(ctx context.Context, id string) (Fact, error)
}

// Fact represents the data about an fact.
type Fact struct {
	entity.Fact
}

// CreateFactRequest represents an fact creation request.
type CreateFactRequest struct {
	CID    string    `json:"cid"`
	RID    string    `json:"rid"`
	Source string    `json:"source`
	Body   string    `json:"body"`
	IAT    time.Time `json:"iat"`
}

// Validate validates the CreateFactRequest fields.
func (m CreateFactRequest) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.Body, validation.Required, validation.Length(0, 128)),
	)
}

// UpdateFactRequest represents an fact update request.
type UpdateFactRequest struct {
	Body string `json:"body"`
}

// Validate validates the CreateFactRequest fields.
func (m UpdateFactRequest) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.Body, validation.Required, validation.Length(0, 128)),
	)
}

type service struct {
	repo   Repository
	logger log.Logger
}

// NewService creates a new fact service.
func NewService(repo Repository, logger log.Logger) Service {
	return service{repo, logger}
}

// Get returns the fact with the specified the fact ID.
func (s service) Get(ctx context.Context, id string) (Fact, error) {
	fact, err := s.repo.Get(ctx, id)
	if err != nil {
		return Fact{}, err
	}
	return Fact{fact}, nil
}

// Create creates a new fact.
func (s service) Create(ctx context.Context, connection string, req CreateFactRequest) (Fact, error) {
	if err := req.Validate(); err != nil {
		return Fact{}, err
	}
	id := entity.GenerateID()
	now := time.Now()
	err := s.repo.Create(ctx, entity.Fact{
		ID:           id,
		ConnectionID: connection,
		ISS:          "me", // TODO: use current app selfid
		Source:       req.Source,
		Body:         req.Body,
		IAT:          req.IAT,
		CreatedAt:    now,
		UpdatedAt:    now,
	})
	if err != nil {
		return Fact{}, err
	}
	return s.Get(ctx, id)
}

// Update updates the fact with the specified ID.
func (s service) Update(ctx context.Context, id string, req UpdateFactRequest) (Fact, error) {
	if err := req.Validate(); err != nil {
		return Fact{}, err
	}

	fact, err := s.Get(ctx, id)
	if err != nil {
		return fact, err
	}
	fact.Body = req.Body
	fact.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, fact.Fact); err != nil {
		return fact, err
	}
	return fact, nil
}

// Delete deletes the fact with the specified ID.
func (s service) Delete(ctx context.Context, id string) (Fact, error) {
	fact, err := s.Get(ctx, id)
	if err != nil {
		return Fact{}, err
	}
	if err = s.repo.Delete(ctx, id); err != nil {
		return Fact{}, err
	}
	return fact, nil
}

// Count returns the number of facts.
func (s service) Count(ctx context.Context) (int, error) {
	return s.repo.Count(ctx)
}

// Query returns the facts with the specified offset and limit.
func (s service) Query(ctx context.Context, connection string, offset, limit int) ([]Fact, error) {
	items, err := s.repo.Query(ctx, connection, offset, limit)
	if err != nil {
		return nil, err
	}
	result := []Fact{}
	for _, item := range items {
		result = append(result, Fact{item})
	}
	return result, nil
}
