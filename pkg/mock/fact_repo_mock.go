package mock

import (
	"context"
	"database/sql"

	"github.com/joinself/restful-client/internal/entity"
)

type FactRepositoryMock struct {
	Items []entity.Fact
}

func (m FactRepositoryMock) Get(ctx context.Context, connID int, id string) (entity.Fact, error) {
	for _, item := range m.Items {
		if item.ID == id {
			return item, nil
		}
	}
	return entity.Fact{}, sql.ErrNoRows
}

func (m FactRepositoryMock) Count(ctx context.Context, conn int, source, fact string) (int, error) {
	return len(m.Items), nil
}

func (m FactRepositoryMock) Query(ctx context.Context, conn int, source, fact string, offset, limit int) ([]entity.Fact, error) {
	return m.Items, nil
}

func (m *FactRepositoryMock) Create(ctx context.Context, fact entity.Fact) error {
	if fact.Fact == "error" {
		return ErrCRUD
	}
	m.Items = append(m.Items, fact)
	return nil
}

func (m *FactRepositoryMock) Update(ctx context.Context, fact entity.Fact) error {
	if fact.Body == "error" {
		return ErrCRUD
	}
	for i, item := range m.Items {
		if item.ID == fact.ID {
			m.Items[i] = fact
			break
		}
	}
	return nil
}

func (m *FactRepositoryMock) Delete(ctx context.Context, connID int, id string) error {
	for i, item := range m.Items {
		if item.ID == id {
			m.Items[i] = m.Items[len(m.Items)-1]
			m.Items = m.Items[:len(m.Items)-1]
			break
		}
	}
	return nil
}

func (m *FactRepositoryMock) SetStatus(ctx context.Context, connID int, id string, status string) error {
	return nil
}

func (r FactRepositoryMock) FindByRequestID(ctx context.Context, connectionID *int, requestID string) ([]entity.Fact, error) {
	return []entity.Fact{}, nil
}
