package signature

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/joinself/restful-client/internal/entity"
	"github.com/joinself/restful-client/pkg/log"
	"github.com/joinself/restful-client/pkg/support"
	"github.com/joinself/self-go-sdk/documents"
)

// Service encapsulates usecase logic for signatures.
type Service interface {
	Get(ctx context.Context, aID, cID, id string) (ExtSignature, error)
	Query(ctx context.Context, aID, cID string, signaturesSince int, offset, limit int) ([]ExtSignature, error)
	Count(ctx context.Context, aID, cID string, signaturesSince int) (int, error)
	Create(ctx context.Context, appID, connectionID string, input CreateSignatureRequest) (ExtSignature, error)
}

func newSignatureFromEntity(m entity.Signature) ExtSignature {
	return ExtSignature{
		ID:          m.ID,
		Description: m.Description,
		Status:      m.Status,
		Data:        m.Data,
		Signature:   m.Signature,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

// CreateSignatureRequest represents an signature creation request.
type service struct {
	repo   Repository
	runner support.SelfClientGetter
	logger log.Logger
}

// NewService creates a new signature service.
func NewService(repo Repository, runner support.SelfClientGetter, logger log.Logger) Service {
	return service{repo, runner, logger}
}

// Get returns the signature with the specified the signature ID.
func (s service) Get(ctx context.Context, aID, cID, id string) (ExtSignature, error) {
	signature, err := s.repo.Get(ctx, aID, cID, id)
	if err != nil {
		return ExtSignature{}, err
	}
	return newSignatureFromEntity(signature), nil
}

// Create creates a new signature.
func (s service) Create(ctx context.Context, appID, selfID string, req CreateSignatureRequest) (ExtSignature, error) {
	now := time.Now()

	cid := uuid.New().String()
	sig := entity.Signature{
		ID:          cid,
		AppID:       appID,
		SelfID:      selfID,
		Description: req.Description,
		Status:      entity.SIGNATURE_REQUESTED_STATUS,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	err := s.repo.Create(ctx, &sig)
	if err != nil {
		return ExtSignature{}, err
	}
	go func() {
		// Send the signature to the connection.
		err := s.requestSignature(appID, selfID, sig, req)
		if err != nil {
			s.markAsErrored(appID, selfID, sig.ID)
			return
		}
	}()

	return s.Get(ctx, appID, selfID, cid)
}

// Count returns the number of signatures.
func (s service) Count(ctx context.Context, aID, cID string, signaturesSince int) (int, error) {
	return s.repo.Count(ctx, aID, cID, signaturesSince)
}

// Query returns the signatures with the specified offset and limit.
func (s service) Query(ctx context.Context, aID, cID string, signaturesSince int, offset, limit int) ([]ExtSignature, error) {
	items, err := s.repo.Query(ctx, aID, cID, signaturesSince, offset, limit)
	if err != nil {
		return nil, err
	}
	result := []ExtSignature{}
	for _, item := range items {
		result = append(result, newSignatureFromEntity(item))
	}
	return result, nil
}

func (s service) requestSignature(appID, connection string, sig entity.Signature, req CreateSignatureRequest) error {
	client, ok := s.runner.Get(appID)
	if !ok {
		return nil
	}

	input := req.Objects[0].DataURI
	b64data := input[strings.IndexByte(input, ',')+1:]
	content, err := base64.RawStdEncoding.DecodeString(b64data)
	mime := input[strings.IndexByte(input, ':')+1 : strings.Index(input, ";")]

	objects := make([]documents.InputObject, 0)
	objects = append(objects, documents.InputObject{
		Name: req.Description,
		Data: content,
		Mime: mime,
	})

	resp, err := client.DocsService().RequestSignature(connection, "Read and sign this documents", objects)
	if err != nil {
		return err
	}
	r, err := s.repo.Get(context.Background(), appID, connection, sig.ID)
	if err != nil {
		return err
	}

	if resp.Status == "accepted" {
		r.Status = entity.SIGNATURE_ACCEPTED_STATUS
		data, err := json.Marshal(resp.SignedObjects)
		if err == nil {
			// TODO: log the error
			r.Data = data
		}
		r.Signature = resp.Signature
	} else {
		r.Status = entity.SIGNATURE_REJECTED_STATUS
	}
	return s.repo.Update(context.Background(), r)
}

func (s service) markAsErrored(appID, connection, id string) error {
	r, err := s.repo.Get(context.Background(), appID, connection, id)
	if err != nil {
		return err
	}

	r.Status = entity.SIGNATURE_ERRORED_STATUS
	return s.repo.Update(context.Background(), r)
}
