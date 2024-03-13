package mock

import (
	"context"
	"database/sql"

	"github.com/joinself/restful-client/internal/entity"
)

type AccountRepositoryMock struct {
	Items []entity.Account
}

func (m AccountRepositoryMock) Get(ctx context.Context, username, pwd string) (entity.Account, error) {
	for _, item := range m.Items {
		if item.UserName == username && item.Password == pwd {
			return item, nil
		}
	}
	return entity.Account{}, sql.ErrNoRows
}

func (m AccountRepositoryMock) Count(ctx context.Context) (int, error) {
	return len(m.Items), nil
}

func (m AccountRepositoryMock) List(ctx context.Context) ([]entity.Account, error) {
	return []entity.Account{}, nil
}

func (m AccountRepositoryMock) GetByUsername(ctx context.Context, username string) (entity.Account, error) {
	for _, item := range m.Items {
		if item.UserName == username {
			return item, nil
		}
	}
	return entity.Account{}, sql.ErrNoRows
}

func (m *AccountRepositoryMock) Create(ctx context.Context, account entity.Account) error {
	if account.UserName == "error" {
		return ErrCRUD
	}
	m.Items = append(m.Items, account)
	return nil
}

func (m *AccountRepositoryMock) Update(ctx context.Context, account entity.Account) error {
	if account.UserName == "error" {
		return ErrCRUD
	}
	for i, item := range m.Items {
		if item.UserName == account.UserName && item.Password == account.Password {
			m.Items[i] = account
			break
		}
	}
	return nil
}

func (m *AccountRepositoryMock) SetPassword(ctx context.Context, id int, password string) error {
	for i, item := range m.Items {
		if item.ID == id {
			m.Items[i].Password = password
			break
		}
	}
	return nil
}

func (m *AccountRepositoryMock) Delete(ctx context.Context, id int) error {
	for i, item := range m.Items {
		if item.ID == id {
			m.Items[i] = m.Items[len(m.Items)-1]
			m.Items = m.Items[:len(m.Items)-1]
			break
		}
	}
	return nil
}
