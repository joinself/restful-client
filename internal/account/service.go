package account

import (
	"context"
	"errors"
	"time"

	"github.com/joinself/restful-client/internal/entity"
	"github.com/joinself/restful-client/pkg/log"
	"github.com/joinself/self-go-sdk/fact"
)

// Service encapsulates usecase logic for accounts.
type Service interface {
	Get(ctx context.Context, username, password string) (Account, error)
	Create(ctx context.Context, input CreateAccountRequest) (Account, error)
	SetPassword(ctx context.Context, username, password, newPassword string) error
	Delete(ctx context.Context, username string) error
	Count(ctx context.Context) (int, error)
	List(ctx context.Context) []ExtAccount
}

// FactService service to manage sending and receiving fact requests
type FactService interface {
	Request(*fact.FactRequest) (*fact.FactResponse, error)
}

// Account represents the data about an account.
type Account struct {
	entity.Account
}

type service struct {
	repo   Repository
	logger log.Logger
}

// NewService creates a new account service.
func NewService(repo Repository, logger log.Logger) Service {
	return service{repo, logger}
}

// Get returns the account with the specified the account ID.
func (s service) Get(ctx context.Context, username, password string) (Account, error) {
	account, err := s.repo.Get(ctx, username, password)
	if err != nil {
		return Account{}, err
	}
	return Account{account}, nil
}

// Create creates a new account.
func (s service) Create(ctx context.Context, req CreateAccountRequest) (Account, error) {
	_, err := s.repo.GetByUsername(ctx, req.Username)
	if err == nil {
		return Account{}, errors.New("user already exists")
	}

	now := time.Now()
	account := entity.Account{
		UserName:               req.Username,
		Password:               req.Password,
		CreatedAt:              now,
		UpdatedAt:              now,
		RequiresPasswordChange: 1,
	}

	if req.RequiresPasswordChange != nil {
		if *req.RequiresPasswordChange == false {
			account.RequiresPasswordChange = 0
		}
	}

	account.SetResources(req.Resources)
	err = s.repo.Create(ctx, account)
	if err != nil {
		return Account{}, err
	}

	return s.Get(ctx, req.Username, req.Password)
}

// SetPassword updates the password for the given account id.
func (s service) SetPassword(ctx context.Context, username, password, newPassword string) error {
	a, err := s.Get(ctx, username, password)
	if err != nil {
		return err
	}

	return s.repo.SetPassword(ctx, a.ID, newPassword)
}

// Delete deletes the account with the specified ID.
func (s service) Delete(ctx context.Context, username string) error {
	account, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		return err
	}

	if err = s.repo.Delete(ctx, account.ID); err != nil {
		return err
	}
	return nil
}

// Count returns the number of accounts.
func (s service) Count(ctx context.Context) (int, error) {
	return s.repo.Count(ctx)
}

// List lists all the existing accounts.
func (s service) List(ctx context.Context) []ExtAccount {
	accounts, err := s.repo.List(ctx)
	if err != nil {
		return []ExtAccount{}
	}

	output := []ExtAccount{}
	for _, a := range accounts {
		output = append(output, ExtAccount{
			UserName:               a.UserName,
			Resources:              a.Resources,
			RequiresPasswordChange: (a.RequiresPasswordChange != 0),
		})
	}

	return output
}
