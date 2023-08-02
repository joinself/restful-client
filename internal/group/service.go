package group

import (
	"context"
	"errors"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/gofrs/uuid"
	"github.com/joinself/restful-client/internal/connection"
	"github.com/joinself/restful-client/internal/entity"
	"github.com/joinself/restful-client/pkg/log"
	selfsdk "github.com/joinself/self-go-sdk"
)

// Group represents the data about a group.
type Group struct {
	entity.Room
	Members []string `json:"members"`
}

type CreateGroupRequest struct {
	Name    string   `json:"name"`
	Members []string `json:"members"`
}

// Validate validates the CreateConnectionRequest fields.
func (m CreateGroupRequest) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.Name, validation.Required, validation.Length(0, 128)),
		validation.Field(&m.Members, validation.Required, validation.Length(0, 128)),
	)
}

type UpdateGroupRequest struct {
	Name    string   `json:"name"`
	Status  string   `json:"status"`
	Members []string `json:"members"`
}

// Validate validates the CreateConnectionRequest fields.
func (m UpdateGroupRequest) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.Name, validation.Required, validation.Length(0, 128)),
	)
}

type Service interface {
	Query(ctx context.Context, appid string, offset, limit int) ([]Group, error)
	Count(ctx context.Context) (int, error)
	Create(ctx context.Context, appID string, req CreateGroupRequest) (Group, error)
	Get(ctx context.Context, appID string, id int) (Group, error)
	Update(ctx context.Context, appID string, id int, input UpdateGroupRequest) (Group, error)
	Delete(ctx context.Context, appID string, id int) error
}

type service struct {
	repo    Repository
	cRepo   connection.Repository
	logger  log.Logger
	clients map[string]*selfsdk.Client
}

func NewService(repo Repository, cRepo connection.Repository, logger log.Logger, clients map[string]*selfsdk.Client) Service {
	return service{repo, cRepo, logger, clients}
}

func (s service) Query(ctx context.Context, appid string, offset, limit int) ([]Group, error) {
	rooms, err := s.repo.Query(ctx, appid, offset, limit)
	if err != nil {
		return []Group{}, err
	}

	groups := []Group{}
	for _, r := range rooms {
		groups = append(groups, Group{Room: r})
	}

	return groups, nil
}

func (s service) Count(ctx context.Context) (int, error) {
	return s.repo.Count(ctx)
}

func (s service) Create(ctx context.Context, appID string, req CreateGroupRequest) (Group, error) {
	var group Group

	if err := req.Validate(); err != nil {
		return group, err
	}

	gid, err := uuid.NewV6()
	if err != nil {
		return group, errors.New("failed creating group gid")
	}

	// Get all the connections
	var connections []entity.Connection
	for _, m := range req.Members {
		conn, err := s.cRepo.Get(ctx, appID, m)
		if err != nil {
			// Let's create the connection in case it does not exist.
			err = s.cRepo.Create(ctx, entity.Connection{
				AppID:  appID,
				SelfID: m,
			})
			if err != nil {
				return group, err
			}

			return group, errors.New("invalid connection")
		}
		connections = append(connections, conn)
	}

	// Create the groups
	room, err := s.repo.Create(ctx, entity.Room{
		GID:    gid.String(),
		Appid:  appID,
		Name:   req.Name,
		Status: entity.GROUP_CREATED_STATUS,
	})
	if err != nil {
		return group, err
	}

	// Create members
	members := []string{}
	for _, conn := range connections {
		members = append(members, conn.SelfID)
		s.repo.AddMember(ctx, entity.RoomConnection{
			RoomID:       room.ID,
			ConnectionID: conn.ID,
		})
	}
	group = Group{
		Room:    room,
		Members: members,
	}
	go s.invite(appID, group)

	return group, nil
}

func (s service) Get(ctx context.Context, appID string, id int) (Group, error) {
	g, err := s.repo.Get(ctx, appID, id)
	if err != nil {
		return Group{}, err
	}

	members := s.getMemberSelfIDs(ctx, g.ID)
	return Group{Room: g, Members: members}, err
}

func (s service) Update(ctx context.Context, appID string, id int, input UpdateGroupRequest) (Group, error) {
	if err := input.Validate(); err != nil {
		return Group{}, err
	}

	room, err := s.repo.Get(ctx, appID, id)
	if err != nil {
		return Group{}, err
	}

	room.Name = input.Name
	room.Status = input.Status
	if room.Status == entity.GROUP_INVITED_STATUS && input.Status == entity.GROUP_JOINED_STATUS {
		go s.join(appID, room)
	}

	err = s.repo.Update(ctx, room)
	if err != nil {
		return Group{}, err
	}

	existingMembers := []entity.Connection{}
	for _, memberID := range s.repo.MemberIDs(ctx, id) {
		c, err := s.cRepo.GetByID(ctx, memberID)
		if err != nil {
			// You cannot create a group with non existing connections, you should create the connection first.
		}
		existingMembers = append(existingMembers, c)
	}

	// TODO: Remove deleted members
	for _, conn := range existingMembers {
		needsRemoval := true
		for _, m := range input.Members {
			if m == conn.SelfID {
				needsRemoval = false
				break
			}
		}
		if needsRemoval {
			s.repo.RemoveMember(ctx, entity.RoomConnection{
				RoomID:       id,
				ConnectionID: conn.ID,
			})
		}
	}

	// TODO: Add new members.
	for _, m := range input.Members {
		needsAddition := true
		for _, conn := range existingMembers {
			if m == conn.SelfID {
				needsAddition = false
				break
			}
		}
		if needsAddition {
			conn, err := s.cRepo.Get(ctx, appID, m)
			if err != nil {
				err = s.cRepo.Create(ctx, entity.Connection{
					AppID:  appID,
					SelfID: m,
				})
				if err != nil {
					return Group{}, errors.New("invalid connection")
				}
			}
			s.repo.AddMember(ctx, entity.RoomConnection{
				RoomID:       id,
				ConnectionID: conn.ID,
			})
		}
	}

	// TODO: Send the group update to Self.

	return Group{
		Room: room,
	}, nil
}

// Delete deletes a group and all its members.
func (s service) Delete(ctx context.Context, appID string, id int) error {
	// Check if the group belongs to the given app.
	group, err := s.repo.Get(ctx, appID, id)
	if err != nil {
		return errors.New("the given group does not belong to the app " + err.Error())
	}

	// Delete the group members
	err = s.repo.RemoveMembers(ctx, id)
	if err != nil {
		return err
	}

	// Delete the group
	err = s.repo.Delete(ctx, id)

	go s.leave(appID, group)

	return err
}

func (s service) getMemberSelfIDs(ctx context.Context, id int) []string {
	members := []string{}
	for _, m := range s.repo.MemberIDs(ctx, id) {
		conn, err := s.cRepo.GetByID(ctx, m)
		if err != nil {
			continue
		}
		members = append(members, conn.SelfID)
	}

	return members
}

func (s service) invite(appID string, g Group) {
	if _, ok := s.clients[appID]; !ok {
		return
	}

	s.clients[appID].ChatService().Invite(g.GID, g.Name, g.Members)
}

func (s service) leave(appID string, g entity.Room) {
	if _, ok := s.clients[appID]; !ok {
		return
	}

	members := s.getMemberSelfIDs(context.Background(), g.ID)
	s.clients[appID].ChatService().Invite(g.GID, g.Name, members)
}

func (s service) join(appID string, g entity.Room) {
	if _, ok := s.clients[appID]; !ok {
		return
	}
	members := s.getMemberSelfIDs(context.Background(), g.ID)
	s.clients[appID].ChatService().Join(g.GID, members)
}
