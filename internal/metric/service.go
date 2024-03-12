package metric

import (
	"context"

	"github.com/joinself/restful-client/internal/config"
	"github.com/joinself/restful-client/internal/entity"
	"github.com/joinself/restful-client/pkg/log"
	"github.com/joinself/self-go-sdk/fact"
)

// Service encapsulates usecase logic for metrics.
type Service interface {
	Query(ctx context.Context, appid string, offset, limit int, from, to int64) ([]Metric, error)
	Count(ctx context.Context, appid string, from, to int64) (int, error)
}

// FactService service to manage sending and receiving fact requests
type FactService interface {
	Request(*fact.FactRequest) (*fact.FactResponse, error)
}

// Metric represents the data about an metric.
type Metric struct {
	entity.Metric
}

type service struct {
	repo       Repository
	signingKey string
	logger     log.Logger
}

// NewService creates a new metric service.
func NewService(cfg *config.Config, repo Repository, logger log.Logger) Service {
	return service{repo, cfg.JWTSigningKey, logger}
}

// Count returns the number of metrics.
func (s service) Count(ctx context.Context, appid string, from, to int64) (int, error) {
	return s.repo.Count(ctx, appid, from, to)
}

// Query returns the metrics with the specified offset and limit.
func (s service) Query(ctx context.Context, appid string, offset, limit int, from, to int64) ([]Metric, error) {
	items, err := s.repo.Query(ctx, appid, offset, limit, from, to)
	if err != nil {
		return nil, err
	}
	result := []Metric{}
	for _, item := range items {
		result = append(result, Metric{item})
	}
	return result, nil
}
