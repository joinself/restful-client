package request

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	b64 "encoding/base64"

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
	Get(ctx context.Context, appID, id string) (Request, error)
	Create(ctx context.Context, appID string, conn *entity.Connection, input CreateRequest) (Request, error)
	CreateFactsFromResponse(conn entity.Connection, req entity.Request, facts []selffact.Fact) []entity.Fact
}

// RequesterService service to manage sending and receiving request requests
type RequesterService interface {
	Request(*selffact.FactRequest) (*selffact.FactResponse, error)
	GenerateQRCode(req *selffact.QRFactRequest) ([]byte, error)
	GenerateDeepLink(req *selffact.DeepLinkFactRequest) (string, error)
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
	QRCode    string            `json:"qr_code,omitempty"`
	DeepLink  string            `json:"deep_link,omitempty"`
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
	SelfID      string        `json:"connection_self_id"`
	OutOfBand   bool          `json:"out_of_band,omitempty"`
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
	dlCodes map[string]string
}

// NewService creates a new request service.
func NewService(repo Repository, fRepo fact.Repository, atRepo attestation.Repository, logger log.Logger, clients map[string]RequesterService, ws map[string]*webhook.Webhook, dlCodes map[string]string) Service {
	return service{
		repo,
		fRepo,
		atRepo,
		logger,
		clients,
		ws,
		dlCodes,
	}
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
func (s service) Create(ctx context.Context, appID string, connection *entity.Connection, req CreateRequest) (Request, error) {
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
		ID:          id,
		Type:        req.Type,
		Facts:       factsBody,
		Status:      "requested",
		Callback:    req.Callback,
		Description: req.Description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if connection != nil && connection.ID != 0 {
		f.ConnectionID = &connection.ID
	}

	if req.Type == "auth" {
		f.Auth = true
	}

	err = s.repo.Create(ctx, f)
	if err != nil {
		return Request{}, err
	}

	if req.OutOfBand {
		r, err := s.buildSelfFactQRRequest(f)
		if err != nil {
			s.logger.Debug("error building Self Fact Request")
			return Request{}, err
		}
		qrdata, err := s.clients[appID].GenerateQRCode(r)
		if err != nil {
			s.logger.Debug("error generating QR Code")
			return Request{}, err
		}
		link := ""
		dlr, err := s.buildSelfFactDLRequest(f, appID)
		if err == nil {
			link, err = s.clients[appID].GenerateDeepLink(dlr)
		}

		persisted, err := s.Get(ctx, appID, id)
		persisted.QRCode = b64.StdEncoding.EncodeToString(qrdata)
		persisted.DeepLink = link

		return persisted, err
	}

	// Send the message to the connection.
	if connection != nil {
		go s.sendRequest(f, appID, connection.SelfID)
	}

	return s.Get(ctx, appID, id)
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
			conn := entity.Connection{
				ID:     *req.ConnectionID,
				SelfID: selfID,
			}
			s.CreateFactsFromResponse(conn, req, resp.Facts)
		}
		s.markRequestAs(req.ID, "responded")
	}

	// TODO: send the current status to the callback function if exists
	go s.sendCallback(appid, selfID, req)
}

func (s service) sendCallback(appID, selfID string, req entity.Request) {
	resp, err := s.Get(context.Background(), appID, req.ID)
	if err != nil {
		s.logger.Info("error getting request: %v", err)
		return
	}

	err = s.w[appID].Post(webhook.WebhookPayload{
		Type: webhook.TYPE_REQUEST,
		URI:  fmt.Sprintf("/apps/%s/connections/%s/requests/%s", appID, selfID, req.ID),
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

// buildSelfFactQRRequest builds a fact request from a given entity.Request
func (s service) buildSelfFactQRRequest(req entity.Request) (*selffact.QRFactRequest, error) {
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

	r := &selffact.QRFactRequest{
		ConversationID: req.ID,
		Description:    req.Description,
		Facts:          facts,
		Expiry:         time.Minute * 5,
		QRConfig: selffact.QRConfig{
			Size:            400,       // this is optional/defaulted
			BackgroundColor: "#FFFFFF", // this is optional/defaulted
			ForegroundColor: "#000000", // this is optional/defaulted
		},
	}

	if req.Auth {
		r.Auth = true
	}

	return r, nil
}

// buildSelfFactQRRequest builds a fact request from a given entity.Request
func (s service) buildSelfFactDLRequest(req entity.Request, appID string) (*selffact.DeepLinkFactRequest, error) {
	if _, ok := s.dlCodes[appID]; !ok || s.dlCodes[appID] == "" {
		return nil, errors.New("dl code not configured")
	}

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

	r := &selffact.DeepLinkFactRequest{
		ConversationID: req.ID,
		Description:    req.Description,
		Facts:          facts,
		Callback:       s.dlCodes[appID],
		Expiry:         time.Minute * 5,
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

func (s service) CreateFactsFromResponse(conn entity.Connection, req entity.Request, facts []selffact.Fact) []entity.Fact {
	output := []entity.Fact{}
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
			ConnectionID: conn.ID,
			ISS:          conn.SelfID,
			Status:       entity.STATUS_ACCEPTED,
			Fact:         receivedFact.Fact,
			Source:       source,
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		if len(req.ID) > 0 {
			f.RequestID = &req.ID
		}

		err := s.fRepo.Create(context.Background(), f)
		if err != nil {
			s.logger.Errorf("failed creating fact: %v", err)
			continue
		}

		s.createAttestations(id, receivedFact)
		output = append(output, f)
	}
	return output
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
