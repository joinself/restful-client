package connection

import (
	"context"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/joinself/restful-client/internal/entity"
	"github.com/joinself/restful-client/pkg/log"
)

// Service encapsulates usecase logic for connections.
type Service interface {
	Get(ctx context.Context, id string) (Connection, error)
	Query(ctx context.Context, offset, limit int) ([]Connection, error)
	Count(ctx context.Context) (int, error)
	Create(ctx context.Context, input CreateConnectionRequest) (Connection, error)
	Update(ctx context.Context, id string, input UpdateConnectionRequest) (Connection, error)
	Delete(ctx context.Context, id string) (Connection, error)
}

// Connection represents the data about an connection.
type Connection struct {
	entity.Connection
}

// CreateConnectionRequest represents an connection creation request.
type CreateConnectionRequest struct {
	SelfID string `json:"selfid"`
	Name   string `json:"name"`
}

// Validate validates the CreateConnectionRequest fields.
func (m CreateConnectionRequest) Validate() error {
	err := validation.ValidateStruct(&m,
		validation.Field(&m.SelfID, validation.Required, validation.Length(0, 128)),
	)
	if err != nil {
		return err
	}

	return validation.ValidateStruct(&m,
		validation.Field(&m.Name, validation.Required, validation.Length(0, 128)),
	)
}

// UpdateConnectionRequest represents an connection update request.
type UpdateConnectionRequest struct {
	Name string `json:"name"`
}

// Validate validates the CreateConnectionRequest fields.
func (m UpdateConnectionRequest) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.Name, validation.Required, validation.Length(0, 128)),
	)
}

type service struct {
	repo   Repository
	logger log.Logger
}

// NewService creates a new connection service.
func NewService(repo Repository, logger log.Logger) Service {
	return service{repo, logger}
}

// Get returns the connection with the specified the connection ID.
func (s service) Get(ctx context.Context, id string) (Connection, error) {
	connection, err := s.repo.Get(ctx, id)
	if err != nil {
		return Connection{}, err
	}
	return Connection{connection}, nil
}

// Create creates a new connection.
func (s service) Create(ctx context.Context, req CreateConnectionRequest) (Connection, error) {
	if err := req.Validate(); err != nil {
		return Connection{}, err
	}
	id := req.SelfID
	existing, err := s.Get(ctx, id)
	if err == nil {
		return existing, nil
	}

	now := time.Now()
	err = s.repo.Create(ctx, entity.Connection{
		ID:        id,
		Name:      req.Name,
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		return Connection{}, err
	}
	return s.Get(ctx, id)
}

// Update updates the connection with the specified ID.
func (s service) Update(ctx context.Context, id string, req UpdateConnectionRequest) (Connection, error) {
	if err := req.Validate(); err != nil {
		return Connection{}, err
	}

	connection, err := s.Get(ctx, id)
	if err != nil {
		return connection, err
	}
	connection.Name = req.Name
	connection.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, connection.Connection); err != nil {
		return connection, err
	}
	return connection, nil
}

// Delete deletes the connection with the specified ID.
func (s service) Delete(ctx context.Context, id string) (Connection, error) {
	connection, err := s.Get(ctx, id)
	if err != nil {
		return Connection{}, err
	}
	if err = s.repo.Delete(ctx, id); err != nil {
		return Connection{}, err
	}
	return connection, nil
}

// Count returns the number of connections.
func (s service) Count(ctx context.Context) (int, error) {
	return s.repo.Count(ctx)
}

// Query returns the connections with the specified offset and limit.
func (s service) Query(ctx context.Context, offset, limit int) ([]Connection, error) {
	items, err := s.repo.Query(ctx, offset, limit)
	if err != nil {
		return nil, err
	}
	result := []Connection{}
	for _, item := range items {
		result = append(result, Connection{item})
	}
	return result, nil
}
