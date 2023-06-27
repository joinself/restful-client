package request

import (
	"context"
	"encoding/json"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/google/uuid"
	"github.com/joinself/restful-client/internal/attestation"
	"github.com/joinself/restful-client/internal/entity"
	"github.com/joinself/restful-client/internal/fact"
	"github.com/joinself/restful-client/pkg/log"
	selffact "github.com/joinself/self-go-sdk/fact"
)

// Service encapsulates usecase logic for requests.
type Service interface {
	Get(ctx context.Context, appID, id string) (Request, error)
	Create(ctx context.Context, appID, selfID string, connection int, input CreateRequest) (Request, error)
}

// RequesterService service to manage sending and receiving request requests
type RequesterService interface {
	Request(*selffact.FactRequest) (*selffact.FactResponse, error)
}

type RequestResource struct {
	URI string `json:"uri"`
}

// Fact represents the data about an request.
type Request struct {
	ID        string            `json:"id"`
	Type      string            `json:"typ"`
	Facts     []FactRequest     `json:"facts"`
	Auth      bool              `json:"auth,omitempty"`
	Status    string            `json:"status"`
	Resources []RequestResource `json:"resources,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

type FactRequest struct {
	Sources []string `json:"sources,omitempty"`
	Name    string   `json:"name"`
}

// CreateRequest represents an request creation request.
type CreateRequest struct {
	Type  string        `json:"type"`
	Facts []FactRequest `json:"facts"`
}

// Validate validates the CreateFactRequest fields.
func (m CreateRequest) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.Type, validation.Required, validation.Length(0, 128)),
	)
}

type service struct {
	repo    Repository
	fRepo   fact.Repository
	atRepo  attestation.Repository
	logger  log.Logger
	clients map[string]RequesterService
}

// NewService creates a new request service.
func NewService(repo Repository, fRepo fact.Repository, atRepo attestation.Repository, logger log.Logger, clients map[string]RequesterService) Service {
	return service{repo, fRepo, atRepo, logger, clients}
}

// Get returns the request with the specified the request ID.
func (s service) Get(ctx context.Context, appID, id string) (Request, error) {
	request, err := s.repo.Get(ctx, id)
	if err != nil {
		return Request{}, err
	}

	var facts []FactRequest
	err = json.Unmarshal(request.Facts, &facts)
	if err != nil {
		return Request{}, err
	}

	resources := []RequestResource{}
	if request.IsResponded() {
		facts, err := s.fRepo.FindByRequestID(ctx, request.ConnectionID, request.ID)
		if err == nil {
			for _, f := range facts {
				resources = append(resources, RequestResource{
					URI: f.URI(appID),
				})
			}
		}
	}

	return Request{
		ID:        request.ID,
		Type:      request.Type,
		Status:    request.Status,
		Auth:      request.Auth,
		Resources: resources,
		Facts:     facts,
	}, nil
}

// Create creates a new request.
func (s service) Create(ctx context.Context, appID, selfID string, connection int, req CreateRequest) (Request, error) {
	if err := req.Validate(); err != nil {
		return Request{}, err
	}
	id := entity.GenerateID()
	now := time.Now()

	facts := make([]entity.RequestFacts, len(req.Facts))
	for i, f := range req.Facts {
		facts[i] = entity.RequestFacts{
			Sources: f.Sources,
			Name:    f.Name,
		}
	}
	factsBody, err := json.Marshal(facts)
	if err != nil {
		return Request{}, err
	}

	f := entity.Request{
		ID:           id,
		ConnectionID: connection,
		Type:         req.Type,
		Facts:        factsBody,
		Status:       "requested",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if req.Type == "auth" {
		f.Auth = true
	}

	err = s.repo.Create(ctx, f)
	if err != nil {
		return Request{}, err
	}

	// Send the message to the connection.
	go s.sendRequest(f, appID, selfID)

	return s.Get(ctx, appID, id)
}

// sendRequest sends a request to the specified connection through Self Network.
func (s service) sendRequest(req entity.Request, appid, selfid string) {
	if _, ok := s.clients[appid]; !ok {
		s.logger.Debug("skipping as self is not initialized")
		return
	}

	var incomingFacts []entity.RequestFacts
	err := json.Unmarshal(req.Facts, &incomingFacts)
	if err != nil {
		s.logger.Errorf("failed processing response: %v", err)
		return
	}

	facts := make([]selffact.Fact, len(incomingFacts))
	for i, f := range incomingFacts {
		facts[i] = selffact.Fact{
			Fact:    f.Name,
			Sources: f.Sources,
		}
	}

	r := &selffact.FactRequest{
		SelfID:      selfid,
		Description: "info",
		Facts:       facts,
		Expiry:      time.Minute * 5,
	}

	if req.Auth {
		r.Auth = true
	}
	resp, err := s.clients[appid].Request(r)

	if err != nil {
		err = s.repo.SetStatus(context.Background(), req.ID, "rejected")
		if err != nil {
			s.logger.Errorf("failed to update status: %v", err)
		}

		s.logger.Errorf("failed to send request: %v", err)
		return
	}

	if len(resp.Facts) != 1 {
		err = s.repo.SetStatus(context.Background(), req.ID, "errored")
		if err != nil {
			s.logger.Errorf("failed to update status: %v", err)
		}
		s.logger.Errorf("unexpected fact response")
		return
	}

	// fact response on one table or another
	for _, receivedFact := range resp.Facts {
		source := ""
		if len(receivedFact.Sources) > 0 {
			source = receivedFact.Sources[0]
		}

		// Create the received fact.
		id := entity.GenerateID()
		now := time.Now()
		f := entity.Fact{
			ID:           id,
			ConnectionID: req.ConnectionID,
			RequestID:    req.ID,
			ISS:          selfid,
			Status:       "accepted",
			Fact:         receivedFact.Fact,
			Source:       source,
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		err := s.fRepo.Create(context.Background(), f)
		if err != nil {
			continue
		}

		// Create the relative attestations.
		for _, v := range resp.Facts[0].AttestedValues() {
			err = s.atRepo.Create(context.Background(), entity.Attestation{
				ID:     uuid.New().String(),
				Body:   "TODO", // TODO: store body.
				FactID: id,
				Value:  v,
			})
			if err != nil {
				continue
			}
		}
	}

	err = s.repo.SetStatus(context.Background(), req.ID, "responded")
	if err != nil {
		s.logger.Errorf("failed to update status: %v", err)
		return
	}

}
