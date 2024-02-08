package app

import (
	"context"
	"time"

	"github.com/joinself/restful-client/internal/entity"
	"github.com/joinself/restful-client/internal/self"
	"github.com/joinself/restful-client/pkg/log"
	"github.com/joinself/self-go-sdk/fact"
)

// Service encapsulates usecase logic for apps.
type Service interface {
	List(ctx context.Context) []entity.App
	ListByStatus(ctx context.Context, statuses []string) ([]entity.App, error)
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
type service struct {
	repo   Repository
	runner self.Runner
	logger log.Logger
}

// NewService creates a new app service.
func NewService(repo Repository, runner self.Runner, logger log.Logger) Service {
	return service{repo, runner, logger}
}

func (s service) List(ctx context.Context) []entity.App {
	apps, err := s.repo.List(ctx)
	if err != nil {
		return []entity.App{}
	}
	return apps
}

// ListByStatus returns the apps that meet the statuses
func (s service) ListByStatus(ctx context.Context, statuses []string) ([]entity.App, error) {
	return s.repo.ListByStatus(ctx, statuses)
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
	existing, err := s.Get(ctx, req.ID)
	if err == nil {
		return existing, nil
	}

	now := time.Now()
	app := entity.App{
		ID:           req.ID,
		DeviceSecret: req.Secret,
		Name:         req.Name,
		Env:          req.Env,
		Status:       entity.APP_CREATED_STATUS,
		Callback:     req.Callback,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	err = s.repo.Create(ctx, app)
	if err != nil {
		return App{}, err
	}

	// Start the runner.
	s.runner.Run(app)

	return s.Get(ctx, req.ID)
}

// Delete deletes the app with the specified ID.
func (s service) Delete(ctx context.Context, id string) (App, error) {
	app, err := s.Get(ctx, id)
	if err != nil {
		return App{}, err
	}

	// Start the runner.
	s.runner.Stop(id)

	if err = s.repo.Delete(ctx, app.ID); err != nil {
		return App{}, err
	}
	return app, nil
}

// Count returns the number of apps.
func (s service) Count(ctx context.Context) (int, error) {
	return s.repo.Count(ctx)
}
