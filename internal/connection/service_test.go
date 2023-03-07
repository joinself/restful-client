package connection

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/joinself/restful-client/internal/entity"
	"github.com/joinself/restful-client/pkg/log"
	"github.com/stretchr/testify/assert"
)

var errCRUD = errors.New("error crud")

func TestCreateConnectionRequest_Validate(t *testing.T) {
	tests := []struct {
		name      string
		model     CreateConnectionRequest
		wantError bool
	}{
		{"success", CreateConnectionRequest{SelfID: "selfid", Name: "test"}, false},
		{"required", CreateConnectionRequest{Name: ""}, true},
		{"too long", CreateConnectionRequest{Name: "1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.model.Validate()
			assert.Equal(t, tt.wantError, err != nil)
		})
	}
}

func TestUpdateConnectionRequest_Validate(t *testing.T) {
	tests := []struct {
		name      string
		model     UpdateConnectionRequest
		wantError bool
	}{
		{"success", UpdateConnectionRequest{Name: "test"}, false},
		{"required", UpdateConnectionRequest{Name: ""}, true},
		{"too long", UpdateConnectionRequest{Name: "1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.model.Validate()
			assert.Equal(t, tt.wantError, err != nil)
		})
	}
}

func Test_service_CRUD(t *testing.T) {
	logger, _ := log.NewForTest()
	s := NewService(&mockRepository{}, logger, nil)

	ctx := context.Background()

	// initial count
	count, _ := s.Count(ctx)
	assert.Equal(t, 0, count)

	// successful creation
	id := "selfid"
	connection, err := s.Create(ctx, CreateConnectionRequest{SelfID: id, Name: "test"})
	assert.Nil(t, err)
	assert.Equal(t, id, connection.ID)
	assert.Equal(t, "test", connection.Name)
	assert.NotEmpty(t, connection.CreatedAt)
	assert.NotEmpty(t, connection.UpdatedAt)
	count, _ = s.Count(ctx)
	assert.Equal(t, 1, count)

	// validation error in creation
	_, err = s.Create(ctx, CreateConnectionRequest{Name: ""})
	assert.NotNil(t, err)
	count, _ = s.Count(ctx)
	assert.Equal(t, 1, count)

	// unexpected error in creation
	_, err = s.Create(ctx, CreateConnectionRequest{SelfID: "selfid2", Name: "error"})
	assert.Equal(t, errCRUD, err)
	count, _ = s.Count(ctx)
	assert.Equal(t, 1, count)

	_, _ = s.Create(ctx, CreateConnectionRequest{SelfID: "selfid2", Name: "test2"})

	// update
	connection, err = s.Update(ctx, id, UpdateConnectionRequest{Name: "test updated"})
	assert.Nil(t, err)
	assert.Equal(t, "test updated", connection.Name)
	_, err = s.Update(ctx, "none", UpdateConnectionRequest{Name: "test updated"})
	assert.NotNil(t, err)

	// validation error in update
	_, err = s.Update(ctx, id, UpdateConnectionRequest{Name: ""})
	assert.NotNil(t, err)
	count, _ = s.Count(ctx)
	assert.Equal(t, 2, count)

	// unexpected error in update
	_, err = s.Update(ctx, id, UpdateConnectionRequest{Name: "error"})
	assert.Equal(t, errCRUD, err)
	count, _ = s.Count(ctx)
	assert.Equal(t, 2, count)

	// get
	_, err = s.Get(ctx, "none")
	assert.NotNil(t, err)
	connection, err = s.Get(ctx, id)
	assert.Nil(t, err)
	assert.Equal(t, "test updated", connection.Name)
	assert.Equal(t, id, connection.ID)

	// query
	connections, _ := s.Query(ctx, 0, 0)
	assert.Equal(t, 2, len(connections))

	// delete
	_, err = s.Delete(ctx, "none")
	assert.NotNil(t, err)
	connection, err = s.Delete(ctx, id)
	assert.Nil(t, err)
	assert.Equal(t, id, connection.ID)
	count, _ = s.Count(ctx)
	assert.Equal(t, 1, count)
}

type mockRepository struct {
	items []entity.Connection
}

func (m mockRepository) Get(ctx context.Context, id string) (entity.Connection, error) {
	for _, item := range m.items {
		if item.ID == id {
			return item, nil
		}
	}
	return entity.Connection{}, sql.ErrNoRows
}

func (m mockRepository) Count(ctx context.Context) (int, error) {
	return len(m.items), nil
}

func (m mockRepository) Query(ctx context.Context, offset, limit int) ([]entity.Connection, error) {
	return m.items, nil
}

func (m *mockRepository) Create(ctx context.Context, connection entity.Connection) error {
	if connection.Name == "error" {
		return errCRUD
	}
	m.items = append(m.items, connection)
	return nil
}

func (m *mockRepository) Update(ctx context.Context, connection entity.Connection) error {
	if connection.Name == "error" {
		return errCRUD
	}
	for i, item := range m.items {
		if item.ID == connection.ID {
			m.items[i] = connection
			break
		}
	}
	return nil
}

func (m *mockRepository) Delete(ctx context.Context, id string) error {
	for i, item := range m.items {
		if item.ID == id {
			m.items[i] = m.items[len(m.items)-1]
			m.items = m.items[:len(m.items)-1]
			break
		}
	}
	return nil
}
