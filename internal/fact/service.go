package fact

import (
	"context"
	"time"

	"github.com/joinself/restful-client/internal/attestation"
	"github.com/joinself/restful-client/internal/entity"
	"github.com/joinself/restful-client/pkg/log"
	"github.com/joinself/restful-client/pkg/support"
	"github.com/joinself/self-go-sdk/fact"
)

// Service encapsulates usecase logic for facts.
type Service interface {
	Get(ctx context.Context, connectionID int, id string) (Fact, error)
	Query(ctx context.Context, conn int, source, fact string, offset, limit int) ([]Fact, error)
	Count(ctx context.Context, conn int, source, fact string) (int, error)
	Create(ctx context.Context, appID, selfID string, connection int, input CreateFactRequest) error
	Update(ctx context.Context, connID int, id string, input UpdateFactRequest) (Fact, error)
	Delete(ctx context.Context, connID int, id string) error
}

// RequesterService service to manage sending and receiving fact requests
type IssuerService interface {
	Issue(selfID string, facts []fact.FactToIssue, viewers []string) error
}

// Fact represents the data about an fact.
type Fact struct {
	entity.Fact
	Attestations []entity.Attestation `json:"attestations"`
}

type FactToIssue struct {
	Key        string          `json:"key"`
	Value      string          `json:"value"`
	Source     string          `json:"source"`
	Group      *fact.FactGroup `json:"group,omitempty"`
	Type       string          `json:"type,omitempty"`
	ExpTimeout *time.Duration  `json:"exp_timeout,omitempty"`
}

type service struct {
	repo   Repository
	atRepo attestation.Repository
	runner support.SelfClientGetter
	logger log.Logger
}

// NewService creates a new fact service.
func NewService(repo Repository, atRepo attestation.Repository, runner support.SelfClientGetter, logger log.Logger) Service {
	return service{repo, atRepo, runner, logger}
}

// Get returns the fact with the specified the fact ID.
func (s service) Get(ctx context.Context, connectionID int, id string) (Fact, error) {
	fact, err := s.repo.Get(ctx, connectionID, id)
	if err != nil {
		return Fact{}, err
	}

	// Get the attestation for the fact.
	attestations, err := s.atRepo.Query(ctx, fact.ID, 0, 1000)
	if err != nil {
		return Fact{}, err
	}

	return Fact{
		Fact:         fact,
		Attestations: attestations,
	}, nil
}

// Create creates a new fact.
func (s service) Create(ctx context.Context, appID, selfID string, connection int, req CreateFactRequest) error {
	// Issue the fact and send it to the user
	s.issueFact(req, appID, selfID)

	return nil
}

// Update updates the fact with the specified ID.
func (s service) Update(ctx context.Context, connID int, id string, req UpdateFactRequest) (Fact, error) {
	fact, err := s.Get(ctx, connID, id)
	if err != nil {
		return fact, err
	}
	fact.Body = req.Body
	fact.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, fact.Fact); err != nil {
		return fact, err
	}
	return fact, nil
}

// Delete deletes the fact with the specified ID.
func (s service) Delete(ctx context.Context, connID int, id string) error {
	if err := s.repo.Delete(ctx, connID, id); err != nil {
		return err
	}
	return nil
}

// Count returns the number of facts.
func (s service) Count(ctx context.Context, conn int, source, fact string) (int, error) {
	return s.repo.Count(ctx, conn, source, fact)
}

// Query returns the facts with the specified offset and limit.
func (s service) Query(ctx context.Context, conn int, source, fact string, offset, limit int) ([]Fact, error) {
	items, err := s.repo.Query(ctx, conn, source, fact, offset, limit)
	if err != nil {
		return nil, err
	}
	result := []Fact{}
	for _, item := range items {
		// Get the attestation for the fact.
		attestations, err := s.atRepo.Query(ctx, item.ID, 0, 1000)
		if err != nil {
			continue
		}

		result = append(result, Fact{
			Fact:         item,
			Attestations: attestations,
		})
	}
	return result, nil
}

// issueFact issues a new fact and sends it to the hwe
func (s service) issueFact(f CreateFactRequest, appid, selfid string) {
	client, ok := s.runner.Get(appid)
	if !ok {
		s.logger.Debug("skipping as self is not initialized")
		return
	}

	fi := []fact.FactToIssue{}
	for _, fa := range f.Facts {
		nf := fact.FactToIssue{
			Key:        fa.Key,
			Value:      fa.Value,
			Source:     fa.Source,
			Type:       fa.Type,
			ExpTimeout: fa.ExpTimeout,
		}

		if fa.Group != nil {
			nf.Group = &fact.FactGroup{
				Name: fa.Group.Name,
				Icon: fa.Group.Icon,
			}
		}

		fi = append(fi, nf)
	}

	client.FactService().Issue(selfid, fi, []string{})
}
