package mock

import (
	"context"
	"database/sql"
	"errors"

	"github.com/joinself/restful-client/internal/entity"
)

type RequestRepositoryMock struct {
	Items []entity.Request
}

func (m RequestRepositoryMock) Get(ctx context.Context, id string) (entity.Request, error) {
	for _, item := range m.Items {
		if item.ID == id {
			return item, nil
		}
	}
	return entity.Request{}, sql.ErrNoRows
}

func (m RequestRepositoryMock) Count(ctx context.Context) (int, error) {
	return len(m.Items), nil
}

func (m RequestRepositoryMock) Query(ctx context.Context, connection, lasRequestID, offset, limit int) ([]entity.Request, error) {
	return m.Items, nil
}

func (m *RequestRepositoryMock) Create(ctx context.Context, request entity.Request) error {
	if request.Description == "error" {
		return ErrCRUD
	}
	m.Items = append(m.Items, request)
	return nil
}

func (m *RequestRepositoryMock) Update(ctx context.Context, request entity.Request) error {
	if request.Description == "error" {
		return ErrCRUD
	}
	for i, item := range m.Items {
		if item.ID == request.ID {
			m.Items[i] = request
			break
		}
	}
	return nil
}

func (m *RequestRepositoryMock) Delete(ctx context.Context, id string) error {
	for i, item := range m.Items {
		if item.ID == id {
			m.Items[i] = m.Items[len(m.Items)-1]
			m.Items = m.Items[:len(m.Items)-1]
			return nil
		}
	}
	return errors.New("not found")
}

func (m *RequestRepositoryMock) SetStatus(ctx context.Context, id, status string) error {
	// TODO: implement this
	return nil
}
