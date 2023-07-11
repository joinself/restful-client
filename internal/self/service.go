package self

import (
	"context"
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
)

// Service interface for self service.
type Service interface {
	Run()
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
	s.onChatMessageHook()
	s.onConnectionRequestHook()
}

func (s *service) onChatMessageHook() {
	if s.client == nil {
		return
	}

	s.client.ChatService().OnMessage(func(cm *chat.Message) {
		// Get connection or create one.
		c, err := s.getOrCreateConnection(cm.ISS)
		if err != nil {
			s.logger.With(context.Background(), "self").Info("error creating connection " + err.Error())
			return
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
			return
		}

		s.callBackClient(msg)
	})
}

func (s *service) callBackClient(msg entity.Message) {
	if s.callbackURL == "" {
		return
	}

	m, err := s.mRepo.Get(context.Background(), msg.ID)
	if err != nil {
		s.logger.With(context.Background(), "self").Info("error retrieving message " + err.Error())
		return
	}

	err = webhook.Post(s.callbackURL, m)
	if err != nil {
		s.logger.With(context.Background(), "self").Info(err.Error())
	}
}

func (s *service) onConnectionRequestHook() {
	if s.client == nil {
		return
	}

	s.client.ChatService().OnConnection(func(iss, status string) {
		parts := strings.Split(iss, ":")
		if len(parts) > 0 {
			iss = parts[0]
		}
		_, err := s.getOrCreateConnection(iss)
		if err != nil {
			s.logger.With(context.Background(), "self").Info("error creating connection " + err.Error())
			return
		}
	})
}

func (s *service) getOrCreateConnection(selfID string) (entity.Connection, error) {
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
