package self

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/joinself/restful-client/internal/connection"
	"github.com/joinself/restful-client/internal/entity"
	"github.com/joinself/restful-client/internal/fact"
	"github.com/joinself/restful-client/internal/message"
	"github.com/joinself/restful-client/internal/metric"
	"github.com/joinself/restful-client/internal/request"
	"github.com/joinself/restful-client/pkg/helper"
	"github.com/joinself/restful-client/pkg/log"
	"github.com/joinself/restful-client/pkg/support"
	"github.com/joinself/restful-client/pkg/webhook"
	selfsdk "github.com/joinself/self-go-sdk"
	"github.com/joinself/self-go-sdk/chat"
	selffact "github.com/joinself/self-go-sdk/fact"
	"github.com/joinself/self-go-sdk/messaging"
)

// Service interface for self service.
type Service interface {
	Run() error
	Stop()
	Get() *selfsdk.Client
	Poster() webhook.Poster
	processFactsQueryResp(body []byte, payload map[string]interface{}) error
	processChatMessage(payload map[string]interface{}) error
	processConnectionResp(payload map[string]interface{}) error
	processIncomingMessage(m *messaging.Message)
	processChatMessageRead(payload map[string]interface{}) error
	processChatMessageDelivered(payload map[string]interface{}) error
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
	MetricRepo     metric.Repository
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
	metRepo  metric.Repository
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
		metRepo:  c.MetricRepo,
		logger:   c.Logger,
		selfID:   c.SelfClient.SelfAppID(),
		rService: c.RequestService,
		w:        c.Poster,
	}
	s.SetupHooks()

	return &s
}

// Run executes the background self listerners.
func (s *service) Run() error {
	s.logger.With(context.Background()).Info("starting self client")
	const maxRetries = 5

	for i := 0; i < maxRetries; i++ {
		err := s.client.Start()
		if err == nil {
			s.logger.With(context.Background()).Info("connection successful")
			return nil
		}

		s.logger.With(context.Background()).Info("Connection failed with %s", err.Error())
		nextRetry := time.Duration(math.Pow(2, float64(i))) * time.Second
		s.logger.With(context.Background()).Info("Retry in %v seconds...\n", nextRetry/time.Second)
		time.Sleep(nextRetry)
	}

	return errors.New("could not start the app")
}

func (s *service) Stop() {
	s.client.Stop()
}

func (s *service) Get() *selfsdk.Client {
	return s.client.Get()
}

func (s *service) Poster() webhook.Poster {
	return s.w
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
	println(" ----> " + payload["typ"].(string))

	switch payload["typ"].(string) {
	case "chat.message":
		_ = s.processChatMessage(payload)

	case "identities.connections.resp":
		_ = s.processConnectionResp(payload)

	case "identities.facts.query.resp":
		_ = s.processFactsQueryResp(m.Payload, payload)

	case "identities.facts.issue":
		_ = s.processIssuedFacts(m.Payload, payload)

	case "chat.message.read":
		// TODO: do something
		_ = s.processChatMessageRead(payload)

	case "chat.voice.setup":
		_ = s.processChatVoiceSetup(payload)

	case "chat.voice.start":
		_ = s.processChatVoiceStart(payload)

	case "chat.voice.accept":
		_ = s.processChatVoiceAccept(payload)

	case "chat.voice.stop":
		_ = s.processChatVoiceStop(payload)

	case "chat.voice.busy":
		_ = s.processChatVoiceBusy(payload)

	}
}

func (s *service) processIssuedFacts(body []byte, payload map[string]interface{}) error {
	metrics, err := parseIncomingMetrics(payload)
	if err != nil {
		s.logger.Error("failed parsing incomming metrics")
	}

	for _, m := range metrics {
		m.AppID = s.selfID
		s.metRepo.Upsert(context.Background(), m)
	}

	return nil
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

	req, err := s.rRepo.GetByID(context.Background(), resp.CID)
	if err != nil {
		req = entity.Request{
			ConnectionID: &conn.ID,
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
	createdFacts := s.rService.CreateFactsFromResponse(conn, req, facts)
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
		},
		Payload: payload,
	})
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

func (s *service) processChatMessageRead(payload map[string]interface{}) error {
	cids := payload["cids"].([]interface{})
	if len(cids) == 0 {
		return errors.New("invalid cids received")
	}

	// Get connection or create one.
	c, err := s.getOrCreateConnection(payload["iss"].(string), "-")
	if err != nil {
		s.logger.With(context.Background(), "self").Info("error creating connection " + err.Error())
		return err
	}

	m, err := s.mRepo.Get(context.Background(), c.ID, cids[0].(string))
	m.Read = true
	return s.mRepo.Update(context.Background(), m)
}

func (s *service) processChatMessageDelivered(payload map[string]interface{}) error {
	cids := payload["cids"].([]interface{})
	if len(cids) == 0 {
		return errors.New("invalid cids received")
	}

	// Get connection or create one.
	c, err := s.getOrCreateConnection(payload["iss"].(string), "-")
	if err != nil {
		s.logger.With(context.Background(), "self").Info("error creating connection " + err.Error())
		return err
	}

	m, err := s.mRepo.Get(context.Background(), c.ID, cids[0].(string))
	m.Received = true
	return s.mRepo.Update(context.Background(), m)
}

func (s *service) processChatVoiceSetup(payload map[string]interface{}) error {
	return s.w.Post(webhook.WebhookPayload{
		Type:    webhook.TYPE_VOICE_SETUP,
		URI:     "",
		Payload: payload,
	})
}

func (s *service) processChatVoiceStart(payload map[string]interface{}) error {
	return s.w.Post(webhook.WebhookPayload{
		Type:    webhook.TYPE_VOICE_START,
		URI:     "",
		Payload: payload,
	})
}

func (s *service) processChatVoiceAccept(payload map[string]interface{}) error {
	return s.w.Post(webhook.WebhookPayload{
		Type:    webhook.TYPE_VOICE_ACCEPT,
		URI:     "",
		Payload: payload,
	})
}

func (s *service) processChatVoiceStop(payload map[string]interface{}) error {
	return s.w.Post(webhook.WebhookPayload{
		Type:    webhook.TYPE_VOICE_STOP,
		URI:     "",
		Payload: payload,
	})
}

func (s *service) processChatVoiceBusy(payload map[string]interface{}) error {
	return s.w.Post(webhook.WebhookPayload{
		Type:    webhook.TYPE_VOICE_BUSY,
		URI:     "",
		Payload: payload,
	})
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
