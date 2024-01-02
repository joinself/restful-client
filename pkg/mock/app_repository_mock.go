package mock

import (
	"context"
	"database/sql"

	"github.com/joinself/restful-client/internal/entity"
)

type AppRepositoryMock struct {
	Items []entity.App
}

func (m AppRepositoryMock) Get(ctx context.Context, id string) (entity.App, error) {
	for _, item := range m.Items {
		if item.ID == id {
			return item, nil
		}
	}
	return entity.App{}, sql.ErrNoRows
}

func (m AppRepositoryMock) Count(ctx context.Context) (int, error) {
	return len(m.Items), nil
}

func (m *AppRepositoryMock) Create(ctx context.Context, app entity.App) error {
	if app.ID == "error" {
		return ErrCRUD
	}
	m.Items = append(m.Items, app)
	return nil
}

func (m *AppRepositoryMock) Update(ctx context.Context, app entity.App) error {
	if app.ID == "error" {
		return ErrCRUD
	}
	for i, item := range m.Items {
		if item.ID == app.ID {
			m.Items[i] = app
			break
		}
	}
	return nil
}

func (m *AppRepositoryMock) Delete(ctx context.Context, id string) error {
	for i, item := range m.Items {
		if item.ID == id {
			m.Items[i] = m.Items[len(m.Items)-1]
			m.Items = m.Items[:len(m.Items)-1]
			break
		}
	}
	return nil
}
