package request

import (
	"context"
	"encoding/json"
	"time"

	b64 "encoding/base64"

	"github.com/google/uuid"
	"github.com/joinself/restful-client/internal/attestation"
	"github.com/joinself/restful-client/internal/entity"
	"github.com/joinself/restful-client/internal/fact"
	"github.com/joinself/restful-client/pkg/log"
	"github.com/joinself/restful-client/pkg/support"
	selffact "github.com/joinself/self-go-sdk/fact"
)

// Service encapsulates usecase logic for requests.
type Service interface {
	Get(ctx context.Context, appID, id string) (ExtRequest, error)
	Create(ctx context.Context, appID string, conn *entity.Connection, input CreateRequest) (ExtRequest, error)
	CreateFactsFromResponse(conn entity.Connection, req entity.Request, facts []selffact.Fact) []entity.Fact
	SetRunner(runner support.SelfClientGetter)
}

// RequesterService service to manage sending and receiving request requests
type RequesterService interface {
	Request(*selffact.FactRequest) (*selffact.FactResponse, error)
	GenerateQRCode(req *selffact.QRFactRequest) ([]byte, error)
	GenerateDeepLink(req *selffact.DeepLinkFactRequest) (string, error)
}

type service struct {
	repo   Repository
	fRepo  fact.Repository
	atRepo attestation.Repository
	runner support.SelfClientGetter
	logger log.Logger
}

// NewService creates a new request service.
func NewService(repo Repository, fRepo fact.Repository, atRepo attestation.Repository, logger log.Logger) Service {
	return &service{
		repo:   repo,
		fRepo:  fRepo,
		atRepo: atRepo,
		logger: logger,
	}
}

func (s *service) SetRunner(runner support.SelfClientGetter) {
	s.runner = runner
}

// Get returns the request with the specified the request ID.
func (s service) Get(ctx context.Context, appID, id string) (ExtRequest, error) {
	request, err := s.repo.Get(ctx, appID, id)
	if err != nil {
		return ExtRequest{}, err
	}

	var facts []FactRequest
	err = json.Unmarshal(request.Facts, &facts)
	if err != nil {
		return ExtRequest{}, err
	}

	resources := []ExtResource{}
	if request.IsResponded() || request.IsOutOfBand() {
		facts, err := s.fRepo.FindByRequestID(ctx, request.ConnectionID, request.ID)
		if err == nil {
			for _, f := range facts {
				resources = append(resources, ExtResource{
					ID:           f.ID,
					ConnectionID: f.ISS,
				})
			}
		}
	}

	return ExtRequest{
		ID:        request.ID,
		Status:    request.Status,
		Resources: resources,
	}, nil
}

// Create creates a new request.
func (s service) Create(ctx context.Context, appID string, connection *entity.Connection, req CreateRequest) (ExtRequest, error) {
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
		return ExtRequest{}, err
	}

	f := entity.Request{
		ID:          id,
		AppID:       appID,
		Type:        req.Type,
		Facts:       factsBody,
		Status:      "requested",
		Callback:    req.Callback,
		Description: req.Description,
		OutOfBand:   req.OutOfBand,
		AllowedFor:  time.Duration(req.AllowedFor),
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
		return ExtRequest{}, err
	}

	if req.OutOfBand {
		r, err := s.buildSelfFactQRRequest(f)
		if err != nil {
			s.logger.Debug("error building Self Fact Request")
			return ExtRequest{}, err
		}

		client, ok := s.runner.Get(appID)
		if !ok {
			s.logger.Debug("client %s not found", appID)
			return ExtRequest{}, err
		}

		qrdata, err := client.FactService().GenerateQRCode(r)
		if err != nil {
			s.logger.Debug("error generating QR Code")
			return ExtRequest{}, err
		}
		link := ""
		dlr, err := s.buildSelfFactDLRequest(f, appID)
		if err == nil {
			link, err = client.FactService().GenerateDeepLink(dlr)
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
	client, ok := s.runner.Get(appid)
	if !ok {
		s.logger.Debug("skipping as self is not initialized")
		return
	}

	// Build a valid Self Fact Request from the given entity.
	r, err := s.buildSelfFactRequestAsync(selfID, req)
	if err != nil {
		s.logger.Debug("error building Self Fact Request")
		return
	}

	// Send the request.
	err = client.FactService().RequestAsync(r)
	if err != nil {
		s.markRequestAs(req.ID, entity.STATUS_ERRORED)
	}
}

// buildSelfFactRequestAsync builds a fact request from a given entity.Request
func (s service) buildSelfFactRequestAsync(selfID string, req entity.Request) (*selffact.FactRequestAsync, error) {
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

	r := &selffact.FactRequestAsync{
		CID:         req.ID,
		SelfID:      selfID,
		Description: req.Description,
		Facts:       facts,
		Expiry:      time.Minute * 5,
		AllowedFor:  req.AllowedFor,
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
	// FIXME: dlCodes must be stored on the database, so consumed through the app_repository instead
	dlCode := "default_non_working_dl_code"

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
		Callback:       dlCode,
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
