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
	"github.com/joinself/restful-client/pkg/webhook"
	selffact "github.com/joinself/self-go-sdk/fact"
)

// Service encapsulates usecase logic for requests.
type Service interface {
	Get(ctx context.Context, appID, connID, id string) (Request, error)
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
	Type        string        `json:"type"`
	Facts       []FactRequest `json:"facts"`
	Description string        `json:"description"`
	Callback    string        `json:"callback"`
}

// Validate validates the CreateRequest fields.
func (m CreateRequest) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.Type, validation.Required, validation.Length(0, 128)),
		validation.Field(&m.Type, validation.In("auth", "fact")),
	)
}

type service struct {
	repo    Repository
	fRepo   fact.Repository
	atRepo  attestation.Repository
	logger  log.Logger
	clients map[string]RequesterService
	w       map[string]*webhook.Webhook
}

// NewService creates a new request service.
func NewService(repo Repository, fRepo fact.Repository, atRepo attestation.Repository, logger log.Logger, clients map[string]RequesterService, ws map[string]*webhook.Webhook) Service {
	return service{
		repo,
		fRepo,
		atRepo,
		logger,
		clients,
		ws,
	}
}

// Get returns the request with the specified the request ID.
func (s service) Get(ctx context.Context, appID, connID, id string) (Request, error) {
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
					URI: f.URI(appID, connID),
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
		Callback:     req.Callback,
		Description:  req.Description,
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

	return s.Get(ctx, appID, selfID, id)
}

// sendRequest sends a request to the specified connection through Self Network.
func (s service) sendRequest(req entity.Request, appid, selfID string) {
	// Check if the self is initialized.
	if _, ok := s.clients[appid]; !ok {
		s.logger.Debug("skipping as self is not initialized")
		return
	}

	// Build a valid Self Fact Request from the given entity.
	r, err := s.buildSelfFactRequest(selfID, req)
	if err != nil {
		s.logger.Debug("error building Self Fact Request")
		return
	}

	// Send the request.
	resp, err := s.clients[appid].Request(r)
	if err != nil {
		s.markRequestAs(req.ID, entity.STATUS_ERRORED)
	} else if resp.Status == "rejected" {
		s.markRequestAs(req.ID, entity.STATUS_REJECTED)
	} else if len(resp.Facts) != 1 && req.Type == "fact" {
		s.markRequestAs(req.ID, entity.STATUS_REJECTED)
	} else {
		// Save the received facts.
		if req.Type == "fact" || req.Type == "auth" {

			s.createFacts(selfID, req, resp.Facts)
		}
		s.markRequestAs(req.ID, "responded")
	}

	// TODO: send the current status to the callback function if exists
	go s.sendCallback(appid, selfID, req)
}

func (s service) sendCallback(appID, selfID string, req entity.Request) {
	resp, err := s.Get(context.Background(), appID, selfID, req.ID)
	if err != nil {
		s.logger.Info("error getting request: %v", err)
		return
	}

	err = s.w[appID].Post(webhook.WebhookPayload{
		Type: webhook.TYPE_REQUEST,
		URI:  req.URI(appID, selfID),
		Data: resp,
	})
	if err != nil {
		s.logger.Info(err)
	}
}

// buildSelfFactRequest builds a fact request from a given entity.Request
func (s service) buildSelfFactRequest(selfID string, req entity.Request) (*selffact.FactRequest, error) {
	var incomingFacts []entity.RequestFacts
	err := json.Unmarshal(req.Facts, &incomingFacts)
	if err != nil {
		s.logger.Errorf("failed processing response: %v", err)
		return nil, err
	}

	facts := make([]selffact.Fact, len(incomingFacts))
	for i, f := range incomingFacts {
		facts[i] = selffact.Fact{
			Fact:    f.Name,
			Sources: f.Sources,
		}
	}

	r := &selffact.FactRequest{
		SelfID:      selfID,
		Description: req.Description,
		Facts:       facts,
		Expiry:      time.Minute * 5,
	}

	if req.Auth {
		r.Auth = true
	}

	return r, nil
}

func (s service) markRequestAs(id, status string) {
	err := s.repo.SetStatus(context.Background(), id, status)
	if err != nil {
		s.logger.Errorf("failed to update status: %v", err)
	}
}

func (s service) createFacts(selfID string, req entity.Request, facts []selffact.Fact) {
	for _, receivedFact := range facts {
		// Create the received fact.
		id := entity.GenerateID()
		now := time.Now()

		source := ""
		if len(receivedFact.Sources) > 0 {
			source = receivedFact.Sources[0]
		}

		f := entity.Fact{
			ID:           id,
			ConnectionID: req.ConnectionID,
			RequestID:    req.ID,
			ISS:          selfID,
			Status:       entity.STATUS_ACCEPTED,
			Fact:         receivedFact.Fact,
			Source:       source,
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		err := s.fRepo.Create(context.Background(), f)
		if err != nil {
			s.logger.Errorf("failed creating fact: %v", err)
			continue
		}

		s.createAttestations(id, receivedFact)
	}
}

func (s service) createAttestations(id string, fact selffact.Fact) {
	// Create the relative attestations.
	now := time.Now()
	for _, v := range fact.AttestedValues() {
		err := s.atRepo.Create(context.Background(), entity.Attestation{
			ID:        uuid.New().String(),
			Body:      "TODO", // TODO: store body.
			FactID:    id,
			Value:     v,
			CreatedAt: now,
			UpdatedAt: now,
		})
		if err != nil {
			s.logger.Errorf("failed creating attestation: %v", err)
			continue
		}
	}
}
