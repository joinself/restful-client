package connection

import (
	"context"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/joinself/restful-client/internal/entity"
	"github.com/joinself/restful-client/pkg/log"
	"github.com/joinself/self-go-sdk/fact"
)

// Service encapsulates usecase logic for connections.
type Service interface {
	Get(ctx context.Context, appid, selfid string) (Connection, error)
	Query(ctx context.Context, appid string, offset, limit int) ([]Connection, error)
	Count(ctx context.Context) (int, error)
	Create(ctx context.Context, appid string, input CreateConnectionRequest) (Connection, error)
	Update(ctx context.Context, appid, selfid string, input UpdateConnectionRequest) (Connection, error)
	Delete(ctx context.Context, appid, selfid string) (Connection, error)
}

// FactService service to manage sending and receiving fact requests
type FactService interface {
	Request(*fact.FactRequest) (*fact.FactResponse, error)
}

// Connection represents the data about an connection.
type Connection struct {
	entity.Connection
}

// CreateConnectionRequest represents an connection creation request.
type CreateConnectionRequest struct {
	SelfID string `json:"selfid"`
}

// Validate validates the CreateConnectionRequest fields.
func (m CreateConnectionRequest) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.SelfID, validation.Required, validation.Length(0, 128)),
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
	repo    Repository
	logger  log.Logger
	clients map[string]FactService
}

// NewService creates a new connection service.
func NewService(repo Repository, logger log.Logger, clients map[string]FactService) Service {
	return service{repo, logger, clients}
}

// Get returns the connection with the specified the connection ID.
func (s service) Get(ctx context.Context, appid, selfid string) (Connection, error) {
	connection, err := s.repo.Get(ctx, appid, selfid)
	if err != nil {
		return Connection{}, err
	}
	return Connection{connection}, nil
}

// Create creates a new connection.
func (s service) Create(ctx context.Context, appid string, req CreateConnectionRequest) (Connection, error) {
	if err := req.Validate(); err != nil {
		return Connection{}, err
	}
	selfid := req.SelfID
	existing, err := s.Get(ctx, appid, selfid)
	if err == nil {
		return existing, nil
	}

	now := time.Now()
	err = s.repo.Create(ctx, entity.Connection{
		SelfID:    selfid,
		AppID:     appid,
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		return Connection{}, err
	}

	go s.requestPublicInfo(appid, selfid)

	return s.Get(ctx, appid, selfid)
}

// Update updates the connection with the specified ID.
func (s service) Update(ctx context.Context, appid, selfid string, req UpdateConnectionRequest) (Connection, error) {
	if err := req.Validate(); err != nil {
		return Connection{}, err
	}

	connection, err := s.Get(ctx, appid, selfid)
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
func (s service) Delete(ctx context.Context, appid, selfid string) (Connection, error) {
	connection, err := s.Get(ctx, appid, selfid)
	if err != nil {
		return Connection{}, err
	}

	if err = s.repo.Delete(ctx, connection.ID); err != nil {
		return Connection{}, err
	}
	return connection, nil
}

// Count returns the number of connections.
func (s service) Count(ctx context.Context) (int, error) {
	return s.repo.Count(ctx)
}

// Query returns the connections with the specified offset and limit.
func (s service) Query(ctx context.Context, appid string, offset, limit int) ([]Connection, error) {
	items, err := s.repo.Query(ctx, appid, offset, limit)
	if err != nil {
		return nil, err
	}
	result := []Connection{}
	for _, item := range items {
		result = append(result, Connection{item})
	}
	return result, nil
}

func (s service) requestPublicInfo(appid, selfid string) {
	if _, ok := s.clients[appid]; !ok {
		s.logger.Debug("skipping as self is not initialized")
		return
	}

	resp, err := s.clients[appid].Request(&fact.FactRequest{
		SelfID:      selfid,
		Description: "info",
		Facts:       []fact.Fact{{Fact: fact.FactDisplayName, Sources: []string{fact.SourceUserSpecified}}},
		Expiry:      time.Minute * 5,
	})
	if err != nil {
		s.logger.Errorf("failed to request public info: %v", err)
		return
	}

	if len(resp.Facts) != 1 {
		s.logger.Errorf("unexpected fact response")
		return
	}

	connection, err := s.Get(context.Background(), appid, selfid)
	if err != nil {
		s.logger.Errorf("unexpected fact response")
		return
	}
	values := resp.Facts[0].AttestedValues()
	if len(values) != 1 {
		s.logger.Errorf("unexpected fact response")
		return
	}

	connection.Name = values[0]

	if err := s.repo.Update(context.Background(), connection.Connection); err != nil {
		s.logger.Errorf("unexpected fact response")
		return
	}
}
