package message

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/joinself/restful-client/internal/entity"
	"github.com/joinself/restful-client/pkg/log"
	"github.com/joinself/restful-client/pkg/support"
	"github.com/joinself/self-go-sdk/chat"
)

// Service encapsulates usecase logic for messages.
type Service interface {
	Get(ctx context.Context, connectionID int, jti string) (Message, error)
	Query(ctx context.Context, connection int, messagesSince int, offset, limit int) ([]Message, error)
	Count(ctx context.Context, connectionID, messagesSince int) (int, error)
	Create(ctx context.Context, appID, connectionID string, connection int, input CreateMessageRequest) (Message, error)
	Update(ctx context.Context, appID string, connectionID int, selfID string, jti string, req UpdateMessageRequest) (Message, error)
	Delete(ctx context.Context, connectionID int, jti string) error
	MarkAsRead(ctx context.Context, appID, connection, jti string) error
	MarkAsReceived(ctx context.Context, appID, connection, jti string) error
}

// Message represents the data about an message.
type Message struct {
	ID           string    `json:"id"`
	ConnectionID string    `json:"connection_id"`
	CID          string    `json:"cid"`
	RID          string    `json:"rid"`
	Body         string    `json:"body"`
	IAT          time.Time `json:"iat"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func newMessageFromEntity(m entity.Message) Message {
	return Message{
		ID:           m.JTI,
		ConnectionID: m.ISS,
		CID:          m.CID,
		RID:          m.RID,
		Body:         m.Body,
		IAT:          m.IAT,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

// CreateMessageRequest represents an message creation request.
type service struct {
	repo   Repository
	runner support.SelfClientGetter
	logger log.Logger
}

// NewService creates a new message service.
func NewService(repo Repository, runner support.SelfClientGetter, logger log.Logger) Service {
	return service{repo, runner, logger}
}

// Get returns the message with the specified the message ID.
func (s service) Get(ctx context.Context, connectionID int, jti string) (Message, error) {
	message, err := s.repo.Get(ctx, connectionID, jti)
	if err != nil {
		return Message{}, err
	}
	return newMessageFromEntity(message), nil
}

// Create creates a new message.
func (s service) Create(ctx context.Context, appID, selfID string, connection int, req CreateMessageRequest) (Message, error) {
	now := time.Now()

	cid := uuid.New().String()
	jti := uuid.New().String()
	msg := entity.Message{
		ISS:          "me",
		ConnectionID: connection,
		CID:          cid,
		JTI:          jti,
		// RID:          req.RID,
		Body: req.Body,
		// IAT:          req.IAT,
		IAT:       now,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Send the message to the connection.
	m, err := s.sendMessage(appID, selfID, req.Body)
	if err != nil {
		return Message{}, err
	}
	if m != nil {
		msg.JTI = m.JTI
	}

	err = s.repo.Create(ctx, &msg)
	if err != nil {
		return Message{}, err
	}

	return s.Get(ctx, connection, msg.JTI)
}

// Update updates the message with the specified ID.
func (s service) Update(ctx context.Context, appID string, connectionID int, selfID string, jti string, req UpdateMessageRequest) (Message, error) {
	message, err := s.repo.Get(ctx, connectionID, jti)
	if err != nil {
		return Message{}, err
	}

	message.Body = req.Body
	message.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, message); err != nil {
		return newMessageFromEntity(message), err
	}

	s.updateMessage(appID, selfID, message.JTI, req.Body)

	return newMessageFromEntity(message), nil
}

// Delete deletes the message with the specified ID.
func (s service) Delete(ctx context.Context, connectionID int, jti string) error {
	return s.repo.Delete(ctx, connectionID, jti)
}

// Count returns the number of messages.
func (s service) Count(ctx context.Context, connectionID, messagesSince int) (int, error) {
	return s.repo.Count(ctx, connectionID, messagesSince)
}

// Query returns the messages with the specified offset and limit.
func (s service) Query(ctx context.Context, connection int, messagesSince int, offset, limit int) ([]Message, error) {
	items, err := s.repo.Query(ctx, connection, messagesSince, offset, limit)
	if err != nil {
		return nil, err
	}
	result := []Message{}
	for _, item := range items {
		result = append(result, newMessageFromEntity(item))
	}
	return result, nil
}

// MarkAsRead marks the given message as read.
func (s service) MarkAsRead(ctx context.Context, appID, connection, jti string) error {
	client, ok := s.runner.Get(appID)
	if !ok {
		return nil
	}
	client.ChatService().Read([]string{connection}, []string{jti}, "")
	return nil
}

// MarkAsRead marks the given message as received.
func (s service) MarkAsReceived(ctx context.Context, appID, connection, jti string) error {
	client, ok := s.runner.Get(appID)
	if !ok {
		return nil
	}
	client.ChatService().Delivered([]string{connection}, []string{jti}, "")
	return nil
}

func (s service) sendMessage(appID, connection string, body string) (*chat.Message, error) {
	client, ok := s.runner.Get(appID)
	if !ok {
		return nil, nil
	}

	return client.ChatService().Message([]string{connection}, body)
}

func (s service) updateMessage(appID, connection, jti, body string) {
	client, ok := s.runner.Get(appID)
	if !ok {
		return
	}

	client.ChatService().Edit(
		[]string{connection},
		jti,
		body,
		"",
	)
}
