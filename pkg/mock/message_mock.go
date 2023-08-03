package mock

import (
	"context"
	"database/sql"
	"errors"

	"github.com/joinself/restful-client/internal/entity"
)

type MessageRepositoryMock struct {
	Items []entity.Message
}

func (m MessageRepositoryMock) Get(ctx context.Context, id string) (entity.Message, error) {
	for _, item := range m.Items {
		if item.JTI == id {
			return item, nil
		}
	}
	return entity.Message{}, sql.ErrNoRows
}

func (m MessageRepositoryMock) Count(ctx context.Context) (int, error) {
	return len(m.Items), nil
}

func (m MessageRepositoryMock) Query(ctx context.Context, connection, lasMessageID, offset, limit int) ([]entity.Message, error) {
	return m.Items, nil
}

func (m *MessageRepositoryMock) Create(ctx context.Context, message *entity.Message) error {
	if message.Body == "error" {
		return ErrCRUD
	}
	m.Items = append(m.Items, *message)
	return nil
}

func (m *MessageRepositoryMock) Update(ctx context.Context, message entity.Message) error {
	if message.Body == "error" {
		return ErrCRUD
	}
	for i, item := range m.Items {
		if item.ID == message.ID {
			m.Items[i] = message
			break
		}
	}
	return nil
}

func (m *MessageRepositoryMock) Delete(ctx context.Context, id string) error {
	for i, item := range m.Items {
		if item.JTI == id {
			m.Items[i] = m.Items[len(m.Items)-1]
			m.Items = m.Items[:len(m.Items)-1]
			return nil
		}
	}
	return errors.New("not found")
}
