package mock

import (
	"context"
	"database/sql"
	"errors"

	"github.com/joinself/restful-client/internal/entity"
)

type SignatureRepositoryMock struct {
	Items []entity.Signature
}

func (m SignatureRepositoryMock) Get(ctx context.Context, appID, selfID, id string) (entity.Signature, error) {
	for _, item := range m.Items {
		if item.ID == id {
			return item, nil
		}
	}
	return entity.Signature{}, sql.ErrNoRows
}

func (m SignatureRepositoryMock) Count(ctx context.Context, aID, cID string, signaturesSince int) (int, error) {
	return len(m.Items), nil
}

func (m SignatureRepositoryMock) Query(ctx context.Context, aID, cID string, signaturesSince int, offset, limit int) ([]entity.Signature, error) {
	return m.Items, nil
}

func (m *SignatureRepositoryMock) Create(ctx context.Context, signature *entity.Signature) error {
	if signature.Description == "error" {
		return ErrCRUD
	}
	m.Items = append(m.Items, *signature)
	return nil
}

func (m *SignatureRepositoryMock) Update(ctx context.Context, signature entity.Signature) error {
	if signature.Description == "error" {
		return ErrCRUD
	}
	for i, item := range m.Items {
		if item.ID == signature.ID {
			m.Items[i] = signature
			break
		}
	}
	return nil
}

func (m *SignatureRepositoryMock) Delete(ctx context.Context, appID, selfID, id string) error {
	for i, item := range m.Items {
		if item.ID == id {
			m.Items[i] = m.Items[len(m.Items)-1]
			m.Items = m.Items[:len(m.Items)-1]
			return nil
		}
	}
	return errors.New("not found")
}
