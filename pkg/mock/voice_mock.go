package mock

import (
	"context"
	"database/sql"
	"errors"

	"github.com/joinself/restful-client/internal/entity"
)

type VoiceRepositoryMock struct {
	Items []entity.Call
}

func (m VoiceRepositoryMock) Get(ctx context.Context, appID, selfID, id string) (entity.Call, error) {
	for _, item := range m.Items {
		if item.CallID == id {
			return item, nil
		}
	}
	return entity.Call{}, sql.ErrNoRows
}

func (m *VoiceRepositoryMock) Create(ctx context.Context, call *entity.Call) error {
	if call.CallID == "error" {
		return ErrCRUD
	}
	m.Items = append(m.Items, *call)
	return nil
}

func (m *VoiceRepositoryMock) Update(ctx context.Context, call entity.Call) error {
	if call.CallID == "error" {
		return ErrCRUD
	}
	for i, item := range m.Items {
		if item.ID == call.ID {
			m.Items[i] = call
			break
		}
	}
	return nil
}

func (m *VoiceRepositoryMock) Delete(ctx context.Context, appID, selfID, id string) error {
	for i, item := range m.Items {
		if item.CallID == id {
			m.Items[i] = m.Items[len(m.Items)-1]
			m.Items = m.Items[:len(m.Items)-1]
			return nil
		}
	}
	return errors.New("not found")
}
