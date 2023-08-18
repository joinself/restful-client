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
	"github.com/joinself/restful-client/internal/request"
	"github.com/joinself/restful-client/pkg/helper"
	"github.com/joinself/restful-client/pkg/log"
	"github.com/joinself/restful-client/pkg/support"
	"github.com/joinself/restful-client/pkg/webhook"
	"github.com/joinself/self-go-sdk/chat"
	selffact "github.com/joinself/self-go-sdk/fact"
	"github.com/joinself/self-go-sdk/messaging"
)

// Service interface for self service.
type Service interface {
	Run()
	processFactsQueryResp(body []byte, payload map[string]interface{}) error
	processChatMessage(payload map[string]interface{}) error
	processConnectionResp(payload map[string]interface{}) error
	processIncomingMessage(m *messaging.Message)
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

type Config struct {
	SelfClient     support.SelfClient
	ConnectionRepo connection.Repository
	FactRepo       fact.Repository
	MessageRepo    message.Repository
	RequestRepo    request.Repository
	Logger         log.Logger
	Poster         webhook.Poster
	RequestService request.Service
}
type service struct {
	client   support.SelfClient
	cRepo    connection.Repository
	fRepo    fact.Repository
	mRepo    message.Repository
	rRepo    request.Repository
	logger   log.Logger
	selfID   string
	w        webhook.Poster
	rService request.Service
}

// NewService creates a new fact service.
func NewService(c Config) Service {
	s := service{
		client:   c.SelfClient,
		cRepo:    c.ConnectionRepo,
		fRepo:    c.FactRepo,
		mRepo:    c.MessageRepo,
		rRepo:    c.RequestRepo,
		logger:   c.Logger,
		selfID:   c.SelfClient.SelfAppID(),
		rService: c.RequestService,
		w:        c.Poster,
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

	s.client.MessagingService().Subscribe("*", s.processIncomingMessage)
}

func (s *service) processIncomingMessage(m *messaging.Message) {
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

	case "identities.facts.query.resp":
		_ = s.processFactsQueryResp(m.Payload, payload)

	}
}

func (s *service) processFactsQueryResp(body []byte, payload map[string]interface{}) error {
	var resp struct {
		*selffact.FactResponse
		CID string `json:"cid"`
	}
	iss := payload["iss"].(string)
	err := json.Unmarshal(body, &resp)
	if err != nil {
		return err
	}
	facts := resp.Facts

	conn, err := s.getOrCreateConnection(iss, "-")
	if err != nil {
		s.logger.With(context.Background(), "self").Info("error creating connection " + err.Error())
		return err
	}

	req, err := s.rRepo.Get(context.Background(), resp.CID)
	if err != nil {
		req = entity.Request{
			ConnectionID: conn.ID,
		}
	} else {
		req.Status = resp.Status
		req.UpdatedAt = time.Now()
		err = s.rRepo.Update(context.Background(), req)
		if err != nil {
			s.logger.With(context.Background(), "self").Info("error updating request " + err.Error())
			return err
		}
	}
	createdFacts := s.rService.CreateFactsFromResponse(iss, req, facts)
	// Return the created facts entity URI.
	for i, _ := range createdFacts {
		createdFacts[i].URL = createdFacts[i].URI(s.selfID)
	}

	// Callback the client webhook
	return s.w.Post(webhook.WebhookPayload{
		Type: webhook.TYPE_FACT_RESPONSE,
		URI:  "",
		Data: entity.Response{
			Facts: createdFacts,
		}})
}

func (s *service) processConnectionResp(payload map[string]interface{}) error {
	iss := payload["iss"].(string)
	parts := strings.Split(iss, ":")
	if len(parts) > 0 {
		iss = parts[0]
	}

	// TODO: we still need to figure out how to manage connection profile image.
	name := "-"
	if data, ok := payload["data"].(map[string]string); ok {
		name = data["name"]
	}

	conn, err := s.getOrCreateConnection(iss, name)
	if err != nil {
		s.logger.With(context.Background(), "self").Info("error creating connection " + err.Error())
		return err
	}

	return s.w.Post(webhook.WebhookPayload{
		Type: webhook.TYPE_CONNECTION,
		URI:  fmt.Sprintf("/apps/%s/connections/%s", s.selfID, conn.SelfID),
		Data: conn})
}

func (s *service) processChatMessage(payload map[string]interface{}) error {
	var cs *chat.Service
	if scs := s.client.ChatService(); scs != nil {
		cs = scs.(*chat.Service)
	}
	cm := chat.NewMessage(cs, []string{payload["aud"].(string)}, payload)

	// Get connection or create one.
	c, err := s.getOrCreateConnection(cm.ISS, "-")
	if err != nil {
		s.logger.With(context.Background(), "self").Info("error creating connection " + err.Error())
		return err
	}

	// Create the input message.
	msg := entity.Message{
		ConnectionID: c.ID,
		ISS:          cm.ISS,
		JTI:          cm.JTI,
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

	return s.w.Post(webhook.WebhookPayload{
		Type: webhook.TYPE_MESSAGE,
		URI:  fmt.Sprintf("/apps/%s/connections/%s/messages/%s", s.selfID, c.SelfID, msg.JTI),
		Data: msg})
}

func (s *service) getOrCreateConnection(selfID, name string) (entity.Connection, error) {
	selfID = helper.FlattenSelfID(selfID)
	c, err := s.cRepo.Get(context.Background(), s.selfID, selfID)
	if err == nil {
		return c, nil
	}

	return s.createConnection(selfID, name)
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
