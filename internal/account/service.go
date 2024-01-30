package account

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/joinself/restful-client/internal/entity"
	"github.com/joinself/restful-client/pkg/log"
	"github.com/joinself/self-go-sdk/fact"
)

// Service encapsulates usecase logic for accounts.
type Service interface {
	Get(ctx context.Context, username, password string) (Account, error)
	Create(ctx context.Context, input CreateAccountRequest) (Account, error)
	Update(ctx context.Context, input UpdateAccountRequest) (Account, error)
	SetPassword(ctx context.Context, username, password, newPassword string) error
	Delete(ctx context.Context, username string) error
	Count(ctx context.Context) (int, error)
}

// FactService service to manage sending and receiving fact requests
type FactService interface {
	Request(*fact.FactRequest) (*fact.FactResponse, error)
}

// Account represents the data about an account.
type Account struct {
	entity.Account
}

// CreateAccountRequest represents an account creation request.
type CreateAccountRequest struct {
	Username  string   `json:"username"`
	Password  string   `json:"password"`
	Resources []string `json:"resources"`
}

// Validate validates the CreateAccountRequest fields.
func (m CreateAccountRequest) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.Username, validation.Required, validation.Length(5, 128)),
		validation.Field(&m.Password, validation.Required, validation.Length(5, 128)),
	)
}

// UpdateAccountRequest represents an account update request.
type UpdateAccountRequest struct {
	Username    string   `json:"username"`
	Password    string   `json:"password"`
	NewPassword string   `json:"new_password"`
	Resources   []string `json:"resources"`
}

// Validate validates the CreateAccountRequest fields.
func (m UpdateAccountRequest) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.Username, validation.Required, validation.Length(5, 128)),
		validation.Field(&m.Password, validation.Required, validation.Length(5, 128)),
	)
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
	if err := req.Validate(); err != nil {
		return Account{}, err
	}
	_, err := s.repo.GetByUsername(ctx, req.Username)
	if err == nil {
		return Account{}, errors.New("user already exists")
	}

	resources, err := json.Marshal(req.Resources)
	if err != nil {
		return Account{}, errors.New("invalid resources")
	}

	now := time.Now()
	err = s.repo.Create(ctx, entity.Account{
		UserName:  req.Username,
		Password:  req.Password,
		Resources: string(resources),
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		return Account{}, err
	}

	return s.Get(ctx, req.Username, req.Password)
}

// Update updates the account with the specified ID.
func (s service) Update(ctx context.Context, req UpdateAccountRequest) (Account, error) {
	if err := req.Validate(); err != nil {
		return Account{}, err
	}

	account, err := s.Get(ctx, req.Username, req.Password)
	if err != nil {
		return account, err
	}

	resources, err := json.Marshal(req.Resources)
	if err != nil {
		return Account{}, errors.New("invalid resources")
	}
	account.UserName = req.Username
	account.Password = req.Password
	account.Resources = string(resources)

	if err := s.repo.Update(ctx, account.Account); err != nil {
		return account, err
	}
	return account, nil
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
