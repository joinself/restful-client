package self

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/joinself/restful-client/internal/connection"
	"github.com/joinself/restful-client/internal/entity"
	"github.com/joinself/restful-client/internal/fact"
	"github.com/joinself/restful-client/internal/message"
	"github.com/joinself/restful-client/pkg/log"
	"github.com/joinself/restful-client/pkg/webhook"
	selfsdk "github.com/joinself/self-go-sdk"
	"github.com/joinself/self-go-sdk/chat"
	"github.com/joinself/self-go-sdk/messaging"
)

const (
	// WEBHOOK_TYPE_MESSAGE webhook type used when a message is received
	WEBHOOK_TYPE_MESSAGE = "message"
	// WEBHOOK_TYPE_FACT_RESPONSE webhook type used when an untracked fact response is received
	WEBHOOK_TYPE_FACT_RESPONSE = "fact_response"
)

// Service interface for self service.
type Service interface {
	Run()
}

// WebhookPayload represents a the payload that will be resent to the
// configured webhook URL if provided.
type WebhookPayload struct {
	// Type is the type of the message.
	Type string `json:"typ"`
	// URI is the URI you can fetch more information about the object on the data field.
	URI string `json"uri"`
	// Data the object to be sent.
	Data interface{} `json:"data"`
}

type service struct {
	client      *selfsdk.Client
	cRepo       connection.Repository
	fRepo       fact.Repository
	mRepo       message.Repository
	logger      log.Logger
	selfID      string
	callbackURL string
}

// NewService creates a new fact service.
func NewService(client *selfsdk.Client, cRepo connection.Repository, fRepo fact.Repository, mRepo message.Repository, callbackURL string, logger log.Logger) Service {
	s := service{
		client:      client,
		cRepo:       cRepo,
		fRepo:       fRepo,
		mRepo:       mRepo,
		logger:      logger,
		selfID:      client.SelfAppID(),
		callbackURL: callbackURL,
	}
	s.SetupHooks()

	return &s
}

// Run executes the background self listerners.
func (s *service) Run() {
	s.logger.With(context.Background(), "self").Info("starting self client")
	err := s.client.Start()
	if err != nil {
		s.logger.With(context.Background(), "self").Error(err.Error())
	}
}

func (s *service) SetupHooks() {
	s.onMessageHook()
}

func (s *service) onMessageHook() {
	if s.client == nil {
		return
	}

	s.client.MessagingService().Subscribe("*", func(m *messaging.Message) {
		var payload map[string]interface{}

		err := json.Unmarshal(m.Payload, &payload)
		if err != nil {
			s.logger.With(context.Background(), "self").Infof("failed to decode message payload: %s", err.Error())
			return
		}

		switch payload["typ"].(string) {
		case "chat.message":
			_ = s.processChatMessage(payload)

		case "identities.connections.resp":
			_ = s.processConnectionResp(payload)

		}
	})
}

func (s *service) processConnectionResp(payload map[string]interface{}) error {
	iss := payload["iss"].(string)
	parts := strings.Split(iss, ":")
	if len(parts) > 0 {
		iss = parts[0]
	}

	conn, err := s.getOrCreateConnection(iss)
	if err != nil {
		s.logger.With(context.Background(), "self").Info("error creating connection " + err.Error())
		return err
	}

	return webhook.Post(s.callbackURL, WebhookPayload{
		Type: WEBHOOK_TYPE_FACT_RESPONSE,
		URI:  fmt.Sprintf("/apps/%s/connections/%s", s.selfID, conn.SelfID),
		Data: conn})
}

func (s *service) processChatMessage(payload map[string]interface{}) error {
	cm := chat.NewMessage(s.client.ChatService(), []string{payload["aud"].(string)}, payload)

	// Get connection or create one.
	c, err := s.getOrCreateConnection(cm.ISS)
	if err != nil {
		s.logger.With(context.Background(), "self").Info("error creating connection " + err.Error())
		return err
	}

	// Create the input message.
	msg := entity.Message{
		ConnectionID: c.ID,
		ISS:          cm.ISS,
		Body:         cm.Body,
		IAT:          time.Now(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err = s.mRepo.Create(context.Background(), &msg)
	if err != nil {
		s.logger.With(context.Background(), "self").Info("error creating message " + err.Error())
		return err
	}

	return webhook.Post(s.callbackURL, WebhookPayload{
		Type: WEBHOOK_TYPE_MESSAGE,
		URI:  fmt.Sprintf("/apps/%s/connections/%s/messages/%d", s.selfID, c.SelfID, msg.ID),
		Data: msg})
}

func (s *service) getOrCreateConnection(selfID string) (entity.Connection, error) {
	selfID = s.flattenSelfID(selfID)
	c, err := s.cRepo.Get(context.Background(), s.selfID, selfID)
	if err == nil {
		return c, nil
	}

	return s.createConnection(selfID, "-")
}

func (s *service) createConnection(selfID, name string) (entity.Connection, error) {
	// Create a connection if it does not exist
	c := entity.Connection{
		Name:      name, // TODO: Send a request to get the user name
		SelfID:    selfID,
		AppID:     s.selfID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := s.cRepo.Create(context.Background(), c)
	if err != nil {
		return c, err
	}

	return s.cRepo.Get(context.Background(), s.selfID, selfID)
}

// TODO: Move this to a helper
func (s *service) flattenSelfID(selfID string) string {
	parts := strings.Split(selfID, ":")
	if len(parts) > 0 {
		return parts[0]
	}

	return selfID
}
