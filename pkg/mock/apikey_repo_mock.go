package mock

import (
	"context"
	"database/sql"

	"github.com/joinself/restful-client/internal/entity"
)

type ApikeyRepositoryMock struct {
	Items []entity.Apikey
}

func (m ApikeyRepositoryMock) Get(ctx context.Context, appid string, id int) (entity.Apikey, error) {
	for _, item := range m.Items {
		if item.AppID == appid && item.ID == id {
			return item, nil
		}
	}
	return entity.Apikey{}, sql.ErrNoRows
}

func (m ApikeyRepositoryMock) Count(ctx context.Context, appID string) (int, error) {
	return len(m.Items), nil
}

func (m ApikeyRepositoryMock) Query(ctx context.Context, appID string, offset, limit int) ([]entity.Apikey, error) {
	return m.Items, nil
}

func (m *ApikeyRepositoryMock) Create(ctx context.Context, apikey *entity.Apikey) error {
	apikey.ID = 99
	if apikey.Name == "error" {
		return ErrCRUD
	}
	m.Items = append(m.Items, *apikey)
	return nil
}

func (m *ApikeyRepositoryMock) Update(ctx context.Context, apikey entity.Apikey) error {
	if apikey.Name == "error" {
		return ErrCRUD
	}
	for i, item := range m.Items {
		if item.ID == apikey.ID && item.AppID == apikey.AppID {
			m.Items[i] = apikey
			break
		}
	}
	return nil
}

func (m *ApikeyRepositoryMock) Delete(ctx context.Context, id int) error {
	for i, item := range m.Items {
		if item.ID == id {
			m.Items[i] = m.Items[len(m.Items)-1]
			m.Items = m.Items[:len(m.Items)-1]
			break
		}
	}
	return nil
}

func (m *ApikeyRepositoryMock) PreloadDeleted(ctx context.Context) error {
	return nil
}
