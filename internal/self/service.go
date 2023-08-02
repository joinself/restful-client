package self

import (
	"context"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/joinself/restful-client/internal/connection"
	"github.com/joinself/restful-client/internal/entity"
	"github.com/joinself/restful-client/internal/fact"
	"github.com/joinself/restful-client/internal/group"
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
	gRepo       group.Repository
	logger      log.Logger
	selfID      string
	callbackURL string
}

// NewService creates a new fact service.
func NewService(client *selfsdk.Client, cRepo connection.Repository, fRepo fact.Repository, mRepo message.Repository, gRepo group.Repository, callbackURL string, logger log.Logger) Service {
	s := service{
		client:      client,
		cRepo:       cRepo,
		fRepo:       fRepo,
		mRepo:       mRepo,
		gRepo:       gRepo,
		logger:      logger,
		selfID:      client.SelfAppID(),
		callbackURL: callbackURL,
	}
	s.SetupHooks()

	return &s
}

// Run executes the background self listerners.
func (s *service) Run() {
	s.logger.With(context.Background()).Info("starting self client")
	err := s.client.Start()
	if err != nil {
		s.logger.With(context.Background()).Error(err.Error())
	}
}

func (s *service) SetupHooks() {
	s.onChatMessageHook()
	s.onConnectionRequestHook()
	s.onGroupInvite()
	s.onGroupJoin()
	s.onGroupLeave()
}

func (s *service) onChatMessageHook() {
	if s.client == nil {
		return
	}

	s.client.ChatService().OnMessage(func(cm *chat.Message) {
		// Create the input message.
		msg := entity.Message{
			ISS:       cm.ISS,
			Body:      cm.Body,
			IAT:       time.Now(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if len(cm.GID) > 0 { // This is a group message.
			// Get the group based on GID
			g, err := s.getGroup(s.selfID, cm.GID)
			if err != nil {
				s.logger.With(context.Background()).Info("error getting group " + err.Error())
				return
			}
			msg.GID = &g.GID
		} else {
			// Get connection or create one.
			c, err := s.getOrCreateConnection(cm.ISS)
			if err != nil {
				s.logger.With(context.Background()).Info("error creating connection " + err.Error())
				return
			}
			msg.ConnectionID = c.ID
		}

		err := s.mRepo.Create(context.Background(), &msg)
		if err != nil {
			s.logger.With(context.Background()).Info("error creating message " + err.Error())
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
		s.logger.With(context.Background()).Info("error retrieving message " + err.Error())
		return
	}

	err = webhook.Post(s.callbackURL, m)
	if err != nil {
		s.logger.With(context.Background()).Info(err.Error())
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
			s.logger.With(context.Background()).Info("error creating connection " + err.Error())
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

func (s *service) getGroup(appID, gid string) (entity.Room, error) {
	return s.gRepo.GetByGID(context.Background(), appID, gid)
}

// Manages invitations to groups.
func (s *service) onGroupInvite() {
	s.client.ChatService().OnInvite(func(g *chat.Group) {
		s.logger.With(context.Background()).Debug("invited to join a group")
		// Go through all members and create connections.
		connections := []entity.Connection{}
		for _, selfID := range g.Members {
			c, err := s.getOrCreateConnection(selfID) // Empty name by now, as is being retrieved
			if err != nil {
				s.logger.With(context.Background()).Info("error creating connection " + err.Error())
				return
			}
			connections = append(connections, c)
		}

		// Crete a group with status invited.
		r, err := s.gRepo.Create(context.Background(), entity.Room{
			Name:   g.Name,
			Status: entity.GROUP_INVITED_STATUS,
			GID:    g.GID,
			Appid:  s.selfID,
		})
		if err != nil {
			s.logger.With(context.Background()).Info("error creating group " + err.Error())
			return
		}

		// Create a group connection.
		for _, c := range connections {
			err = s.gRepo.AddMember(context.Background(), entity.RoomConnection{
				ConnectionID: c.ID,
				RoomID:       r.ID,
			})
			if err != nil {
				s.logger.With(context.Background()).Info("error creating group connection " + err.Error())
				return
			}
		}

		// TODO: make this configurable.
		s.logger.With(context.Background()).Info("automatically joining the group ", g.GID)
		spew.Dump(g.Members)
		// g.Join()
		s.client.ChatService().Join(g.GID, g.Members)
		r.Status = entity.GROUP_JOINED_STATUS
		s.gRepo.Update(context.Background(), r)

		// Send a request to the callback if is setup
	})
}

func (s *service) onGroupJoin() {
	s.client.ChatService().OnJoin(func(iss, gid string) {
		// Someone has joined the group.
		s.logger.With(context.Background()).Debug("group join message received")
		// Get the group from database.
		group, err := s.gRepo.GetByGID(context.Background(), s.selfID, gid)
		if err != nil {
			s.logger.With(context.Background()).Info("error getting group " + err.Error())
			return
		}

		// Add connection if it does not exist
		c, err := s.getOrCreateConnection(iss)
		if err != nil {
			s.logger.With(context.Background()).Info("error creating connection " + err.Error())
			return
		}

		// Update group members.
		members := s.gRepo.MemberIDs(context.Background(), group.ID)
		exists := false
		for _, m := range members {
			if m == c.ID {
				exists = true
				break
			}
		}
		if !exists {
			s.gRepo.AddMember(context.Background(), entity.RoomConnection{
				ConnectionID: c.ID,
				RoomID:       group.ID,
			})
		}
	})
}

func (s *service) onGroupLeave() {
	s.client.ChatService().OnLeave(func(iss, gid string) {
		s.logger.With(context.Background()).Debug("group leave message received")
		// Someone has left the group.
		// Get the group from database.
		group, err := s.gRepo.GetByGID(context.Background(), s.selfID, gid)
		if err != nil {
			s.logger.With(context.Background()).Info("error getting group " + err.Error())
			return
		}

		// Update group members.
		c, err := s.cRepo.Get(context.Background(), s.selfID, iss)
		if err != nil {
			s.logger.With(context.Background()).Info("error getting connection " + err.Error())
			return
		}

		err = s.gRepo.RemoveMember(context.Background(), entity.RoomConnection{
			ConnectionID: c.ID,
			RoomID:       group.ID,
		})
		if err != nil {
			s.logger.With(context.Background()).Info("error removing connection " + err.Error())
		}

		// FIXME: Should we remove the group if you're the last one?
	})
}
