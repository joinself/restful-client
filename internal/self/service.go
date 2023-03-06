package self

import (
	"context"
	"time"

	"github.com/joinself/restful-client/internal/connection"
	"github.com/joinself/restful-client/internal/entity"
	"github.com/joinself/restful-client/internal/fact"
	"github.com/joinself/restful-client/internal/message"
	"github.com/joinself/restful-client/pkg/log"
	selfsdk "github.com/joinself/self-go-sdk"
	"github.com/joinself/self-go-sdk/chat"
)

type Service interface {
}

type service struct {
	client *selfsdk.Client
	cRepo  connection.Repository
	fRepo  fact.Repository
	mRepo  message.Repository
	logger log.Logger
}

// NewService creates a new fact service.
func NewService(client *selfsdk.Client, cRepo connection.Repository, fRepo fact.Repository, mRepo message.Repository, logger log.Logger) Service {
	s := service{
		client: client,
		cRepo:  cRepo,
		fRepo:  fRepo,
		mRepo:  mRepo,
		logger: logger,
	}
	s.SetupHooks()
	client.Start()

	return s
}

func (s *service) SetupHooks() {
	s.onChatMessageHook()
}

func (s *service) onChatMessageHook() {
	if s.client == nil {
		return
	}

	s.client.ChatService().OnMessage(func(cm *chat.Message) {
		// Get connection or create one.
		c, err := s.cRepo.Get(context.Background(), cm.ISS)
		if err != nil {
			// Create a connection if it does not exist
			c := entity.Connection{
				ID:        cm.ISS,
				Name:      "-", // TODO: Send a request to get the user name
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			err = s.cRepo.Create(context.Background(), c)
			if err != nil {
				s.logger.With(context.Background(), "self").Info("error creating connection " + err.Error())
				return
			}
		}

		// Create the input message.
		s.mRepo.Create(context.Background(), &entity.Message{
			ConnectionID: c.ID,
			ISS:          cm.ISS,
			Body:         cm.Body,
			IAT:          time.Now(),
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		})
	})

}
