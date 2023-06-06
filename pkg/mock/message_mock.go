package mock

import (
	"context"
	"database/sql"

	"github.com/joinself/restful-client/internal/entity"
)

type MessageRepositoryMock struct {
	Items []entity.Message
}

func (m MessageRepositoryMock) Get(ctx context.Context, id int) (entity.Message, error) {
	for _, item := range m.Items {
		if item.ID == id {
			return item, nil
		}
	}
	return entity.Message{}, sql.ErrNoRows
}

func (m MessageRepositoryMock) Count(ctx context.Context) (int, error) {
	return len(m.Items), nil
}

func (m MessageRepositoryMock) Query(ctx context.Context, connection string, lasMessageID, offset, limit int) ([]entity.Message, error) {
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

func (m *MessageRepositoryMock) Delete(ctx context.Context, id int) error {
	for i, item := range m.Items {
		if item.ID == id {
			m.Items[i] = m.Items[len(m.Items)-1]
			m.Items = m.Items[:len(m.Items)-1]
			break
		}
	}
	return nil
}
