package message

import (
	"context"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/qiangxue/go-rest-api/internal/entity"
	"github.com/qiangxue/go-rest-api/pkg/log"
)

// Service encapsulates usecase logic for messages.
type Service interface {
	Get(ctx context.Context, id string) (Message, error)
	Query(ctx context.Context, connection string, offset, limit int) ([]Message, error)
	Count(ctx context.Context) (int, error)
	Create(ctx context.Context, connection string, input CreateMessageRequest) (Message, error)
	Update(ctx context.Context, id string, input UpdateMessageRequest) (Message, error)
	Delete(ctx context.Context, id string) (Message, error)
}

// Message represents the data about an message.
type Message struct {
	entity.Message
}

// CreateMessageRequest represents an message creation request.
type CreateMessageRequest struct {
	CID  string    `json:"cid"`
	RID  string    `json:"rid"`
	Body string    `json:"body"`
	IAT  time.Time `json:"iat"`
}

// Validate validates the CreateMessageRequest fields.
func (m CreateMessageRequest) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.Body, validation.Required, validation.Length(0, 128)),
	)
}

// UpdateMessageRequest represents an message update request.
type UpdateMessageRequest struct {
	Body string `json:"body"`
}

// Validate validates the CreateMessageRequest fields.
func (m UpdateMessageRequest) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.Body, validation.Required, validation.Length(0, 128)),
	)
}

type service struct {
	repo   Repository
	logger log.Logger
}

// NewService creates a new message service.
func NewService(repo Repository, logger log.Logger) Service {
	return service{repo, logger}
}

// Get returns the message with the specified the message ID.
func (s service) Get(ctx context.Context, id string) (Message, error) {
	message, err := s.repo.Get(ctx, id)
	if err != nil {
		return Message{}, err
	}
	return Message{message}, nil
}

// Create creates a new message.
func (s service) Create(ctx context.Context, connection string, req CreateMessageRequest) (Message, error) {
	if err := req.Validate(); err != nil {
		return Message{}, err
	}
	id := entity.GenerateID()
	now := time.Now()
	err := s.repo.Create(ctx, entity.Message{
		ID:           id,
		ConnectionID: connection,
		CID:          req.CID,
		RID:          req.RID,
		Body:         req.Body,
		IAT:          req.IAT,
		CreatedAt:    now,
		UpdatedAt:    now,
	})
	if err != nil {
		return Message{}, err
	}
	return s.Get(ctx, id)
}

// Update updates the message with the specified ID.
func (s service) Update(ctx context.Context, id string, req UpdateMessageRequest) (Message, error) {
	if err := req.Validate(); err != nil {
		return Message{}, err
	}

	message, err := s.Get(ctx, id)
	if err != nil {
		return message, err
	}
	message.Body = req.Body
	message.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, message.Message); err != nil {
		return message, err
	}
	return message, nil
}

// Delete deletes the message with the specified ID.
func (s service) Delete(ctx context.Context, id string) (Message, error) {
	message, err := s.Get(ctx, id)
	if err != nil {
		return Message{}, err
	}
	if err = s.repo.Delete(ctx, id); err != nil {
		return Message{}, err
	}
	return message, nil
}

// Count returns the number of messages.
func (s service) Count(ctx context.Context) (int, error) {
	return s.repo.Count(ctx)
}

// Query returns the messages with the specified offset and limit.
func (s service) Query(ctx context.Context, connection string, offset, limit int) ([]Message, error) {
	items, err := s.repo.Query(ctx, connection, offset, limit)
	if err != nil {
		return nil, err
	}
	result := []Message{}
	for _, item := range items {
		result = append(result, Message{item})
	}
	return result, nil
}
