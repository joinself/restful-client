package mock

import (
	"context"
	"database/sql"

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

func (m *RequestRepositoryMock) Create(ctx context.Context, request entity.Request) error {
	if request.Type == "error" {
		return ErrCRUD
	}
	m.Items = append(m.Items, request)
	return nil
}

func (m *RequestRepositoryMock) Update(ctx context.Context, request entity.Request) error {
	if request.Type == "error" {
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
			break
		}
	}
	return nil
}

func (m *RequestRepositoryMock) SetStatus(ctx context.Context, id string, status string) error {
	return nil
}
