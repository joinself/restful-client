package fact

import (
	"context"
	"errors"
	"testing"

	"github.com/joinself/restful-client/pkg/log"
	"github.com/joinself/restful-client/pkg/mock"
	"github.com/stretchr/testify/assert"
)

var errCRUD = errors.New("error crud")

func TestCreateFactRequest_Validate(t *testing.T) {
	tests := []struct {
		name      string
		model     CreateFactRequest
		wantError bool
	}{
		{"success", CreateFactRequest{Fact: "test"}, false},
		{"required", CreateFactRequest{Fact: ""}, true},
		{"too long", CreateFactRequest{Fact: "1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890"}, true},
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
	s := NewService(&mock.FactRepositoryMock{}, &mock.AttestationRepositoryMock{}, logger, nil)

	ctx := context.Background()

	// initial count
	count, _ := s.Count(ctx, 1, "", "")
	assert.Equal(t, 0, count)

	// successful creation
	fact, err := s.Create(ctx, "app", "connection", 1, CreateFactRequest{Fact: "test"})
	assert.Nil(t, err)
	assert.NotEmpty(t, fact.ID)
	id := fact.ID
	assert.Equal(t, "test", fact.Fact.Fact)
	assert.NotEmpty(t, fact.CreatedAt)
	assert.NotEmpty(t, fact.UpdatedAt)
	count, _ = s.Count(ctx, 1, "", "")
	assert.Equal(t, 1, count)

	// validation error in creation
	_, err = s.Create(ctx, "app", "connection", 1, CreateFactRequest{Fact: ""})
	assert.NotNil(t, err)
	count, _ = s.Count(ctx, 1, "", "")
	assert.Equal(t, 1, count)

	// unexpected error in creation
	_, err = s.Create(ctx, "app", "connection", 1, CreateFactRequest{Fact: "error"})
	assert.Equal(t, errCRUD, err)
	count, _ = s.Count(ctx, 1, "", "")
	assert.Equal(t, 1, count)

	_, _ = s.Create(ctx, "app", "connection", 1, CreateFactRequest{Fact: "test2"})

	// update
	fact, err = s.Update(ctx, id, UpdateFactRequest{Body: "test updated"})
	assert.Nil(t, err)
	assert.Equal(t, "test updated", fact.Body)
	_, err = s.Update(ctx, "none", UpdateFactRequest{Body: "test updated"})
	assert.NotNil(t, err)

	// validation error in update
	_, err = s.Update(ctx, id, UpdateFactRequest{Body: ""})
	assert.NotNil(t, err)
	count, _ = s.Count(ctx, 1, "", "")
	assert.Equal(t, 2, count)

	// unexpected error in update
	_, err = s.Update(ctx, id, UpdateFactRequest{Body: "error"})
	assert.Equal(t, errCRUD, err)
	count, _ = s.Count(ctx, 1, "", "")
	assert.Equal(t, 2, count)

	// get
	_, err = s.Get(ctx, "none")
	assert.NotNil(t, err)
	fact, err = s.Get(ctx, id)
	assert.Nil(t, err)
	assert.Equal(t, "test updated", fact.Body)
	assert.Equal(t, id, fact.ID)

	// query
	facts, _ := s.Query(ctx, 1, "", "", 0, 0)
	assert.Equal(t, 2, len(facts))

	// delete
	_, err = s.Delete(ctx, "none")
	assert.NotNil(t, err)
	fact, err = s.Delete(ctx, id)
	assert.Nil(t, err)
	assert.Equal(t, id, fact.ID)
	count, _ = s.Count(ctx, 1, "", "")
	assert.Equal(t, 1, count)
}
