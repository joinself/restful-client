package mock

import (
	"context"
	"database/sql"
	"errors"

	"github.com/joinself/restful-client/internal/entity"
)

var ErrCRUD = errors.New("error crud")

type ConnectionRepositoryMock struct {
	Items []entity.Connection
}

func (m ConnectionRepositoryMock) Get(ctx context.Context, appid, selfid string) (entity.Connection, error) {
	for _, item := range m.Items {
		if item.AppID == appid && item.SelfID == selfid {
			return item, nil
		}
	}
	return entity.Connection{}, sql.ErrNoRows
}

func (m ConnectionRepositoryMock) Count(ctx context.Context) (int, error) {
	return len(m.Items), nil
}

func (m ConnectionRepositoryMock) Query(ctx context.Context, offset, limit int) ([]entity.Connection, error) {
	return m.Items, nil
}

func (m *ConnectionRepositoryMock) Create(ctx context.Context, connection entity.Connection) error {
	if connection.SelfID == "error" {
		return ErrCRUD
	}
	m.Items = append(m.Items, connection)
	return nil
}

func (m *ConnectionRepositoryMock) Update(ctx context.Context, connection entity.Connection) error {
	if connection.Name == "error" {
		return ErrCRUD
	}
	for i, item := range m.Items {
		if item.SelfID == connection.SelfID && item.AppID == connection.AppID {
			m.Items[i] = connection
			break
		}
	}
	return nil
}

func (m *ConnectionRepositoryMock) Delete(ctx context.Context, id int) error {
	for i, item := range m.Items {
		if item.ID == id {
			m.Items[i] = m.Items[len(m.Items)-1]
			m.Items = m.Items[:len(m.Items)-1]
			break
		}
	}
	return nil
}
