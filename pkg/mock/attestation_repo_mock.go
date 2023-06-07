package mock

import (
	"context"
	"database/sql"

	"github.com/joinself/restful-client/internal/entity"
)

type AttestationRepositoryMock struct {
	Items []entity.Attestation
}

func (m AttestationRepositoryMock) Get(ctx context.Context, id string) (entity.Attestation, error) {
	for _, item := range m.Items {
		if item.ID == id {
			return item, nil
		}
	}
	return entity.Attestation{}, sql.ErrNoRows
}

func (m AttestationRepositoryMock) Count(ctx context.Context) (int, error) {
	return len(m.Items), nil
}

func (m AttestationRepositoryMock) Query(ctx context.Context, connection string, offset, limit int) ([]entity.Attestation, error) {
	return m.Items, nil
}

func (m *AttestationRepositoryMock) Create(ctx context.Context, attestation entity.Attestation) error {
	m.Items = append(m.Items, attestation)
	return nil
}
