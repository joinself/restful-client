package apikey

import (
	"context"
	"errors"
	"time"

	"github.com/joinself/restful-client/internal/config"
	"github.com/joinself/restful-client/internal/entity"
	"github.com/joinself/restful-client/pkg/acl"
	"github.com/joinself/restful-client/pkg/log"
	"github.com/joinself/self-go-sdk/fact"
)

// TODO: Make expiration configurable
const GENERATED_TOKEN_EXPIRATION = 999999

// Service encapsulates usecase logic for apikeys.
type Service interface {
	Get(ctx context.Context, appid string, id int) (ApiKey, error)
	Query(ctx context.Context, appid string, offset, limit int) ([]ApiKey, error)
	Count(ctx context.Context, appid string) (int, error)
	Create(ctx context.Context, appid string, input CreateApiKeyRequest, user acl.Identity) (ApiKey, error)
	Update(ctx context.Context, appid string, id int, input UpdateApiKeyRequest) (ApiKey, error)
	Delete(ctx context.Context, appid string, id int) (ApiKey, error)
}

// FactService service to manage sending and receiving fact requests
type FactService interface {
	Request(*fact.FactRequest) (*fact.FactResponse, error)
}

// ApiKey represents the data about an apikey.
type ApiKey struct {
	entity.Apikey
}

type service struct {
	repo       Repository
	signingKey string
	logger     log.Logger
}

// NewService creates a new apikey service.
func NewService(cfg *config.Config, repo Repository, logger log.Logger) Service {
	return service{repo, cfg.JWTSigningKey, logger}
}

// Get returns the apikey with the specified the apikey ID.
func (s service) Get(ctx context.Context, appid string, id int) (ApiKey, error) {
	apikey, err := s.repo.Get(ctx, appid, id)
	if err != nil {
		return ApiKey{}, err
	}
	return ApiKey{apikey}, nil
}

// Create creates a new apikey.
func (s service) Create(ctx context.Context, appid string, req CreateApiKeyRequest, user acl.Identity) (ApiKey, error) {
	now := time.Now()
	ak := entity.Apikey{
		AppID:     appid,
		Name:      req.Name,
		Token:     "",
		CreatedAt: now,
		UpdatedAt: now,
	}
	err := s.repo.Create(ctx, &ak)
	if err != nil {
		return ApiKey{}, err
	}

	tok, err := acl.GenerateJWTToken(entity.User{
		ID:                     user.GetID(),
		Name:                   user.GetName(),
		Admin:                  false,
		RequiresPasswordChange: false,
		Resources:              req.GetResources(appid),
	}, ak.ID, s.signingKey, GENERATED_TOKEN_EXPIRATION)
	if err != nil {
		s.logger.With(ctx).Infof("cannot generate a valid token %v", err)
		return ApiKey{}, errors.New("cannot generate a valid token")
	}

	ak.Token = tok[:3] + "..." + tok[len(tok)-3:]
	err = s.repo.Update(ctx, ak)
	if err != nil {
		s.logger.With(ctx).Infof("cannot generate a valid token %v", err)
		return ApiKey{}, errors.New("cannot generate a valid token")
	}

	rak, err := s.Get(ctx, appid, ak.ID)
	if err != nil {
		s.logger.With(ctx).Infof("error getting api key %s from app %s %v", ak.ID, appid, err)
		return ApiKey{}, err
	}

	rak.Token = tok
	return rak, nil
}

// Update updates the apikey with the specified ID.
func (s service) Update(ctx context.Context, appid string, id int, req UpdateApiKeyRequest) (ApiKey, error) {
	apikey, err := s.Get(ctx, appid, id)
	if err != nil {
		s.logger.With(ctx).Infof("error getting api key %s from app %s %v", id, appid, err)
		return apikey, err
	}
	apikey.Name = req.Name
	apikey.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, apikey.Apikey); err != nil {
		s.logger.With(ctx).Infof("error updating api key %s from app %s %v", apikey.ID, appid, err)
		return apikey, err
	}
	return apikey, nil
}

// Delete deletes the apikey with the specified ID.
func (s service) Delete(ctx context.Context, appid string, id int) (ApiKey, error) {
	apikey, err := s.Get(ctx, appid, id)
	if err != nil {
		s.logger.With(ctx).Infof("error getting api key %s from app %s %v", id, appid, err)
		return ApiKey{}, err
	}

	if err = s.repo.Delete(ctx, apikey.ID); err != nil {
		s.logger.With(ctx).Infof("error deleting api key %s from app %s %v", apikey.ID, appid, err)
		return ApiKey{}, err
	}
	return apikey, nil
}

// Count returns the number of apikeys.
func (s service) Count(ctx context.Context, appid string) (int, error) {
	return s.repo.Count(ctx, appid)
}

// Query returns the apikeys with the specified offset and limit.
func (s service) Query(ctx context.Context, appid string, offset, limit int) ([]ApiKey, error) {
	items, err := s.repo.Query(ctx, appid, offset, limit)
	if err != nil {
		s.logger.With(ctx).Infof("error retrieving apikeys for app %s %v", appid, err)
		return nil, err
	}
	result := []ApiKey{}
	for _, item := range items {
		result = append(result, ApiKey{item})
	}
	return result, nil
}
