package mock

import (
	"context"
	"database/sql"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/joinself/restful-client/internal/entity"
)

type Item struct {
	*entity.Room
	Members []string
}

type GroupRepositoryMock struct {
	Items []Item
}

func (m GroupRepositoryMock) Get(ctx context.Context, appID string, id int) (entity.Room, error) {
	for _, item := range m.Items {
		if item.Appid == appID && item.ID == id {
			// Reset the members
			return *item.Room, nil
		}
	}
	return entity.Room{}, sql.ErrNoRows
}

func (m GroupRepositoryMock) Count(ctx context.Context) (int, error) {
	return len(m.Items), nil
}

func (m GroupRepositoryMock) Query(ctx context.Context, appID string, offset, limit int) ([]entity.Room, error) {
	rooms := []entity.Room{}
	for _, item := range m.Items {
		rooms = append(rooms, *item.Room)
	}
	return rooms, nil
}

func (m *GroupRepositoryMock) Create(ctx context.Context, group entity.Room) (entity.Room, error) {
	if group.Name == "error" {
		return group, ErrCRUD
	}
	rand.Seed(time.Now().UnixNano())
	group.ID = rand.Intn(1000)
	group.CreatedAt = time.Now()
	group.UpdatedAt = time.Now()
	m.Items = append(m.Items, Item{
		Room:    &group,
		Members: []string{},
	})

	return group, nil
}

func (m *GroupRepositoryMock) Update(ctx context.Context, group entity.Room) error {
	if group.Name == "error" {
		return ErrCRUD
	}
	for i, item := range m.Items {
		if item.ID == group.ID && item.Appid == group.Appid {
			m.Items[i] = Item{
				Room:    &group,
				Members: []string{},
			}
			break
		}
	}
	return nil
}

func (m *GroupRepositoryMock) Delete(ctx context.Context, id int) error {
	for i, item := range m.Items {
		if item.ID == id {
			m.Items[i] = m.Items[len(m.Items)-1]
			m.Items = m.Items[:len(m.Items)-1]
			break
		}
	}
	return nil
}

func (m *GroupRepositoryMock) MemberIDs(ctx context.Context, id int) []int {
	// TODO: implement
	members := []int{}
	for i, _ := range m.Items {
		if m.Items[i].ID == id {
			for _, m := range m.Items[i].Members {
				num, err := strconv.Atoi(m)
				if err != nil {
					log.Println(err)
				}
				members = append(members, num)
			}
		}
	}

	return members
}

// AddMember adds a member to the group
func (m *GroupRepositoryMock) AddMember(ctx context.Context, relation entity.RoomConnection) error {
	for i, _ := range m.Items {
		if m.Items[i].ID == relation.RoomID {
			if m.Items[i].Members == nil {
				m.Items[i].Members = []string{
					strconv.Itoa(relation.ConnectionID),
				}
			} else {
				m.Items[i].Members = append(
					m.Items[i].Members,
					strconv.Itoa(relation.ConnectionID))
			}
			return nil
		}
	}
	return nil
}

// RemoveMember removes a member from the group.
func (m *GroupRepositoryMock) RemoveMember(ctx context.Context, relation entity.RoomConnection) error {
	for i, _ := range m.Items {
		if m.Items[i].ID == relation.RoomID {
			members := []string{}
			for _, m := range m.Items[i].Members {
				if m != strconv.Itoa(relation.ConnectionID) {
					members = append(members, m)
				}
			}
			m.Items[i].Members = members
			return nil
		}
	}
	return nil
}

// RemoveMembers removes all members associated to a specific group
func (m *GroupRepositoryMock) RemoveMembers(ctx context.Context, id int) error {
	for i, _ := range m.Items {
		if m.Items[i].ID == id {
			m.Items[i].Members = []string{}
			return nil
		}
	}
	return nil
}
