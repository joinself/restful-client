package fact

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/qiangxue/go-rest-api/internal/entity"
	"github.com/qiangxue/go-rest-api/pkg/log"
	"github.com/stretchr/testify/assert"
)

var errCRUD = errors.New("error crud")

func TestCreateFactRequest_Validate(t *testing.T) {
	tests := []struct {
		name      string
		model     CreateFactRequest
		wantError bool
	}{
		{"success", CreateFactRequest{Body: "test"}, false},
		{"required", CreateFactRequest{Body: ""}, true},
		{"too long", CreateFactRequest{Body: "1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.model.Validate()
			assert.Equal(t, tt.wantError, err != nil)
		})
	}
}

func TestUpdateFactRequest_Validate(t *testing.T) {
	tests := []struct {
		name      string
		model     UpdateFactRequest
		wantError bool
	}{
		{"success", UpdateFactRequest{Body: "test"}, false},
		{"required", UpdateFactRequest{Body: ""}, true},
		{"too long", UpdateFactRequest{Body: "1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890"}, true},
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
	s := NewService(&mockRepository{}, logger)

	ctx := context.Background()

	// initial count
	count, _ := s.Count(ctx)
	assert.Equal(t, 0, count)

	// successful creation
	fact, err := s.Create(ctx, "connection", CreateFactRequest{Body: "test"})
	assert.Nil(t, err)
	assert.NotEmpty(t, fact.ID)
	id := fact.ID
	assert.Equal(t, "test", fact.Body)
	assert.NotEmpty(t, fact.CreatedAt)
	assert.NotEmpty(t, fact.UpdatedAt)
	count, _ = s.Count(ctx)
	assert.Equal(t, 1, count)

	// validation error in creation
	_, err = s.Create(ctx, "connection", CreateFactRequest{Body: ""})
	assert.NotNil(t, err)
	count, _ = s.Count(ctx)
	assert.Equal(t, 1, count)

	// unexpected error in creation
	_, err = s.Create(ctx, "connection", CreateFactRequest{Body: "error"})
	assert.Equal(t, errCRUD, err)
	count, _ = s.Count(ctx)
	assert.Equal(t, 1, count)

	_, _ = s.Create(ctx, "connection", CreateFactRequest{Body: "test2"})

	// update
	fact, err = s.Update(ctx, id, UpdateFactRequest{Body: "test updated"})
	assert.Nil(t, err)
	assert.Equal(t, "test updated", fact.Body)
	_, err = s.Update(ctx, "none", UpdateFactRequest{Body: "test updated"})
	assert.NotNil(t, err)

	// validation error in update
	_, err = s.Update(ctx, id, UpdateFactRequest{Body: ""})
	assert.NotNil(t, err)
	count, _ = s.Count(ctx)
	assert.Equal(t, 2, count)

	// unexpected error in update
	_, err = s.Update(ctx, id, UpdateFactRequest{Body: "error"})
	assert.Equal(t, errCRUD, err)
	count, _ = s.Count(ctx)
	assert.Equal(t, 2, count)

	// get
	_, err = s.Get(ctx, "none")
	assert.NotNil(t, err)
	fact, err = s.Get(ctx, id)
	assert.Nil(t, err)
	assert.Equal(t, "test updated", fact.Body)
	assert.Equal(t, id, fact.ID)

	// query
	facts, _ := s.Query(ctx, "connection", 0, 0)
	assert.Equal(t, 2, len(facts))

	// delete
	_, err = s.Delete(ctx, "none")
	assert.NotNil(t, err)
	fact, err = s.Delete(ctx, id)
	assert.Nil(t, err)
	assert.Equal(t, id, fact.ID)
	count, _ = s.Count(ctx)
	assert.Equal(t, 1, count)
}

type mockRepository struct {
	items []entity.Fact
}

func (m mockRepository) Get(ctx context.Context, id string) (entity.Fact, error) {
	for _, item := range m.items {
		if item.ID == id {
			return item, nil
		}
	}
	return entity.Fact{}, sql.ErrNoRows
}

func (m mockRepository) Count(ctx context.Context) (int, error) {
	return len(m.items), nil
}

func (m mockRepository) Query(ctx context.Context, connection string, offset, limit int) ([]entity.Fact, error) {
	return m.items, nil
}

func (m *mockRepository) Create(ctx context.Context, fact entity.Fact) error {
	if fact.Body == "error" {
		return errCRUD
	}
	m.items = append(m.items, fact)
	return nil
}

func (m *mockRepository) Update(ctx context.Context, fact entity.Fact) error {
	if fact.Body == "error" {
		return errCRUD
	}
	for i, item := range m.items {
		if item.ID == fact.ID {
			m.items[i] = fact
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