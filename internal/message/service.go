package message

import (
	"context"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/google/uuid"
	"github.com/joinself/restful-client/internal/entity"
	"github.com/joinself/restful-client/pkg/log"
	selfsdk "github.com/joinself/self-go-sdk"
	"github.com/joinself/self-go-sdk/chat"
)

// Service encapsulates usecase logic for messages.
type Service interface {
	Get(ctx context.Context, id int) (Message, error)
	Query(ctx context.Context, connection int, messagesSince int, offset, limit int) ([]Message, error)
	Count(ctx context.Context) (int, error)
	Create(ctx context.Context, appID, connectionID string, connection int, input CreateMessageRequest) (Message, error)
	Update(ctx context.Context, appID, connectionID string, id int, req UpdateMessageRequest) (Message, error)
	Delete(ctx context.Context, id int) (Message, error)
}

// Message represents the data about an message.
type Message struct {
	entity.Message
}

// CreateMessageRequest represents an message creation request.
type CreateMessageRequest struct {
	Body string `json:"body"`
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
	repo    Repository
	logger  log.Logger
	clients map[string]*selfsdk.Client
}

// NewService creates a new message service.
func NewService(repo Repository, logger log.Logger, clients map[string]*selfsdk.Client) Service {
	return service{repo, logger, clients}
}

// Get returns the message with the specified the message ID.
func (s service) Get(ctx context.Context, id int) (Message, error) {
	message, err := s.repo.Get(ctx, id)
	if err != nil {
		return Message{}, err
	}
	return Message{message}, nil
}

// Create creates a new message.
func (s service) Create(ctx context.Context, appID, connectionID string, connection int, req CreateMessageRequest) (Message, error) {
	if err := req.Validate(); err != nil {
		return Message{}, err
	}
	now := time.Now()

	cid := uuid.New().String()
	msg := entity.Message{
		ISS:          "me",
		ConnectionID: connection,
		CID:          cid,
		// RID:          req.RID,
		Body: req.Body,
		// IAT:          req.IAT,
		IAT:       now,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Send the message to the connection.
	m, err := s.sendMessage(appID, connectionID, req.Body)
	if err != nil {
		return Message{}, err
	}
	msg.JTI = m.JTI

	err = s.repo.Create(ctx, &msg)
	if err != nil {
		return Message{}, err
	}

	return s.Get(ctx, msg.ID)
}

// Update updates the message with the specified ID.
func (s service) Update(ctx context.Context, appID, connectionID string, id int, req UpdateMessageRequest) (Message, error) {
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

	s.updateMessage(appID, connectionID, message.JTI, req.Body)

	return message, nil
}

// Delete deletes the message with the specified ID.
func (s service) Delete(ctx context.Context, id int) (Message, error) {
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
func (s service) Query(ctx context.Context, connection int, messagesSince int, offset, limit int) ([]Message, error) {
	items, err := s.repo.Query(ctx, connection, messagesSince, offset, limit)
	if err != nil {
		return nil, err
	}
	result := []Message{}
	for _, item := range items {
		result = append(result, Message{item})
	}
	return result, nil
}

func (s service) sendMessage(appID, connection string, body string) (*chat.Message, error) {
	if _, ok := s.clients[appID]; !ok {
		return nil, nil
	}

	return s.clients[appID].ChatService().Message([]string{connection}, body)
}

func (s service) updateMessage(appID, connection, jti, body string) {
	if _, ok := s.clients[appID]; !ok {
		return
	}

	s.clients[appID].ChatService().Edit(
		[]string{connection},
		jti,
		body,
		"",
	)
}
