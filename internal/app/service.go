package app

import (
	"context"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/joinself/restful-client/internal/entity"
	"github.com/joinself/restful-client/pkg/log"
	"github.com/joinself/self-go-sdk/fact"
)

// Service encapsulates usecase logic for apps.
type Service interface {
	Get(ctx context.Context, id string) (App, error)
	Create(ctx context.Context, input CreateAppRequest) (App, error)
	Delete(ctx context.Context, id string) (App, error)
}

// FactService service to manage sending and receiving fact requests
type FactService interface {
	Request(*fact.FactRequest) (*fact.FactResponse, error)
}

// App represents the data about an app.
type App struct {
	entity.App
}

// CreateAppRequest represents an app creation request.
type CreateAppRequest struct {
	ID       string `json:"id"`
	Secret   string `json:"secret"`
	Name     string `json:"name"`
	Env      string `json:"env"`
	Callback string `json:"callback"`
}

// Validate validates the CreateAppRequest fields.
func (m CreateAppRequest) Validate() error {
	return validation.ValidateStruct(&m,
		// TODO: Improve validations
		validation.Field(&m.ID, validation.Required, validation.Length(0, 128)),
		validation.Field(&m.Secret, validation.Required, validation.Length(0, 128)),
		validation.Field(&m.Name, validation.Required, validation.Length(0, 50)),
		validation.Field(&m.Env, validation.Required, validation.Length(0, 10)),
	)
}

type service struct {
	repo   Repository
	logger log.Logger
}

// NewService creates a new app service.
func NewService(repo Repository, logger log.Logger) Service {
	return service{repo, logger}
}

// Get returns the app with the specified the app ID.
func (s service) Get(ctx context.Context, id string) (App, error) {
	app, err := s.repo.Get(ctx, id)
	if err != nil {
		return App{}, err
	}
	return App{app}, nil
}

// Create creates a new app.
func (s service) Create(ctx context.Context, req CreateAppRequest) (App, error) {
	if err := req.Validate(); err != nil {
		return App{}, err
	}
	existing, err := s.Get(ctx, req.ID)
	if err == nil {
		return existing, nil
	}

	now := time.Now()
	err = s.repo.Create(ctx, entity.App{
		ID:           req.ID,
		DeviceSecret: req.Secret,
		Name:         req.Name,
		Env:          req.Env,
		Callback:     req.Callback,
		CreatedAt:    now,
		UpdatedAt:    now,
	})
	if err != nil {
		return App{}, err
	}

	return s.Get(ctx, req.ID)
}

// Delete deletes the app with the specified ID.
func (s service) Delete(ctx context.Context, id string) (App, error) {
	app, err := s.Get(ctx, id)
	if err != nil {
		return App{}, err
	}

	if err = s.repo.Delete(ctx, app.ID); err != nil {
		return App{}, err
	}
	return app, nil
}

// Count returns the number of apps.
func (s service) Count(ctx context.Context) (int, error) {
	return s.repo.Count(ctx)
}
