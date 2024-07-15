package connection

import (
	"context"
	"time"

	"github.com/joinself/restful-client/internal/entity"
	"github.com/joinself/restful-client/pkg/log"
	"github.com/joinself/restful-client/pkg/support"
	"github.com/joinself/self-go-sdk/fact"
)

// Service encapsulates usecase logic for connections.
type Service interface {
	Get(ctx context.Context, appid, selfid string) (Connection, error)
	Query(ctx context.Context, appid string, offset, limit int) ([]Connection, error)
	Count(ctx context.Context, appid string) (int, error)
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

type service struct {
	repo   Repository
	runner support.SelfClientGetter
	logger log.Logger
}

// NewService creates a new connection service.
func NewService(repo Repository, runner support.SelfClientGetter, logger log.Logger) Service {
	return service{repo, runner, logger}
}

// Get returns the connection with the specified the connection ID.
func (s service) Get(ctx context.Context, appid, selfid string) (Connection, error) {
	connection, err := s.repo.Get(ctx, appid, selfid)
	if err != nil {
		s.logger.With(ctx).Infof("cannot get the app %v", err)
		return Connection{}, err
	}
	return Connection{connection}, nil
}

// Create creates a new connection.
func (s service) Create(ctx context.Context, appid string, req CreateConnectionRequest) (Connection, error) {
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
		s.logger.With(ctx).Infof("problem creating the requested app %v", err)
		return Connection{}, err
	}

	go s.requestPublicInfo(appid, selfid)

	return s.Get(ctx, appid, selfid)
}

// Update updates the connection with the specified ID.
func (s service) Update(ctx context.Context, appid, selfid string, req UpdateConnectionRequest) (Connection, error) {
	connection, err := s.Get(ctx, appid, selfid)
	if err != nil {
		s.logger.With(ctx).Infof("cannot get the app %v", err)
		return connection, err
	}
	connection.Name = req.Name
	connection.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, connection.Connection); err != nil {
		s.logger.With(ctx).Infof("problem updating the app %v", err)
		return connection, err
	}
	return connection, nil
}

// Delete deletes the connection with the specified ID.
func (s service) Delete(ctx context.Context, appid, selfid string) (Connection, error) {
	connection, err := s.Get(ctx, appid, selfid)
	if err != nil {
		s.logger.With(ctx).Infof("cannot get the app %v", err)
		return Connection{}, err
	}

	if err = s.repo.Delete(ctx, connection.ID); err != nil {
		s.logger.With(ctx).Infof("problem deleting the app %v", err)
		return Connection{}, err
	}
	return connection, nil
}

// Count returns the number of connections.
func (s service) Count(ctx context.Context, appid string) (int, error) {
	return s.repo.Count(ctx, appid)
}

// Query returns the connections with the specified offset and limit.
func (s service) Query(ctx context.Context, appid string, offset, limit int) ([]Connection, error) {
	items, err := s.repo.Query(ctx, appid, offset, limit)
	if err != nil {
		s.logger.With(ctx).Infof("problem retrieving apps list %v", err)
		return nil, err
	}
	result := []Connection{}
	for _, item := range items {
		result = append(result, Connection{item})
	}
	return result, nil
}

func (s service) requestPublicInfo(appid, selfid string) {
	client, ok := s.runner.Get(appid)
	if !ok {
		s.logger.Debug("skipping as self is not initialized")
		return
	}

	err := client.FactService().RequestAsync(&fact.FactRequestAsync{
		SelfID:      selfid,
		Description: "info",
		Facts:       []fact.Fact{{Fact: fact.FactDisplayName, Sources: []string{fact.SourceUserSpecified}}},
		Expiry:      time.Minute * 5,
	})
	if err != nil {
		s.logger.Warnf("failed to request public info: %v", err)
		return
	}
}
