package fact

import (
	"context"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/joinself/restful-client/internal/entity"
	"github.com/joinself/restful-client/pkg/log"
	"github.com/joinself/self-go-sdk/fact"
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

type FactService interface {
	Request(*fact.FactRequest) (*fact.FactResponse, error)
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
	Fact   string    `json:"fact`
	Body   string    `json:"body"`
	IAT    time.Time `json:"iat"`
}

// Validate validates the CreateFactRequest fields.
func (m CreateFactRequest) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.Fact, validation.Required, validation.Length(0, 128)),
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
	client FactService
}

// NewService creates a new fact service.
func NewService(repo Repository, logger log.Logger, client FactService) Service {
	return service{repo, logger, client}
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

	println("......")
	println(req.Fact)
	println("......")

	f := entity.Fact{
		ID:           id,
		ConnectionID: connection,
		ISS:          "me", // TODO: use current app selfid
		Source:       req.Source,
		Fact:         req.Fact,
		Body:         req.Body,
		IAT:          req.IAT,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	err := s.repo.Create(ctx, f)
	if err != nil {
		return Fact{}, err
	}

	// Send the message to the connection.
	if s.client != nil {
		go s.sendRequest(f)
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

// sendRequest sends a request to the specified connection through Self Network.
func (s service) sendRequest(f entity.Fact) {
	if s.client == nil {
		s.logger.Debug("skipping as self is not initialized")
		return
	}

	resp, err := s.client.Request(&fact.FactRequest{
		SelfID:      f.ConnectionID,
		Description: "info",
		Facts:       []fact.Fact{{Fact: f.Fact, Sources: []string{f.Source}}},
		Expiry:      time.Minute * 5,
	})
	if err != nil {
		err = s.repo.SetStatus(context.Background(), f.ID, "errored")
		if err != nil {
			s.logger.Errorf("failed to update status: %v", err)
		}

		s.logger.Errorf("failed to send request: %v", err)
		return
	}

	if len(resp.Facts) != 1 {
		err = s.repo.SetStatus(context.Background(), f.ID, "errored")
		if err != nil {
			s.logger.Errorf("failed to update status: %v", err)
		}
		s.logger.Errorf("unexpected fact response")
		return
	}

	err = s.repo.SetStatus(context.Background(), f.ID, "received")
	if err != nil {
		s.logger.Errorf("failed to update status: %v", err)
		return
	}

	// Create the relative attestations.
	for _, v := range resp.Facts[0].AttestedValues() {
		println(" - " + v)
	}
}
