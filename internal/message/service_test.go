package message

import (
	"context"
	"errors"
	"testing"

	"github.com/joinself/restful-client/pkg/log"
	"github.com/joinself/restful-client/pkg/mock"
	"github.com/stretchr/testify/assert"
)

var errCRUD = errors.New("error crud")

func TestCreateMessageRequest_Validate(t *testing.T) {
	tests := []struct {
		name      string
		model     CreateMessageRequest
		wantError bool
	}{
		{"success", CreateMessageRequest{Body: "test"}, false},
		{"required", CreateMessageRequest{Body: ""}, true},
		{"too long", CreateMessageRequest{Body: "1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.model.Validate()
			assert.Equal(t, tt.wantError, err != nil)
		})
	}
}

func TestUpdateMessageRequest_Validate(t *testing.T) {
	tests := []struct {
		name      string
		model     UpdateMessageRequest
		wantError bool
	}{
		{"success", UpdateMessageRequest{Body: "test"}, false},
		{"required", UpdateMessageRequest{Body: ""}, true},
		{"too long", UpdateMessageRequest{Body: "1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890"}, true},
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
	s := NewService(&mock.MessageRepositoryMock{}, logger, nil)

	ctx := context.Background()

	// initial count
	count, _ := s.Count(ctx)
	assert.Equal(t, 0, count)

	// successful creation
	message, err := s.Create(ctx, "connection", CreateMessageRequest{Body: "test"})
	assert.Nil(t, err)
	id := message.ID
	assert.Equal(t, "test", message.Body)
	assert.NotEmpty(t, message.CreatedAt)
	assert.NotEmpty(t, message.UpdatedAt)
	count, _ = s.Count(ctx)
	assert.Equal(t, 1, count)

	// validation error in creation
	_, err = s.Create(ctx, "connection", CreateMessageRequest{Body: ""})
	assert.NotNil(t, err)
	count, _ = s.Count(ctx)
	assert.Equal(t, 1, count)

	// unexpected error in creation
	_, err = s.Create(ctx, "connection", CreateMessageRequest{Body: "error"})
	assert.Equal(t, errCRUD, err)
	count, _ = s.Count(ctx)
	assert.Equal(t, 1, count)

	_, _ = s.Create(ctx, "connection", CreateMessageRequest{Body: "test2"})

	// update
	message, err = s.Update(ctx, id, UpdateMessageRequest{Body: "test updated"})
	assert.Nil(t, err)
	assert.Equal(t, "test updated", message.Body)
	_, err = s.Update(ctx, 1, UpdateMessageRequest{Body: "test updated"})
	assert.NotNil(t, err)

	// validation error in update
	_, err = s.Update(ctx, id, UpdateMessageRequest{Body: ""})
	assert.NotNil(t, err)
	count, _ = s.Count(ctx)
	assert.Equal(t, 2, count)

	// unexpected error in update
	_, err = s.Update(ctx, id, UpdateMessageRequest{Body: "error"})
	assert.Equal(t, errCRUD, err)
	count, _ = s.Count(ctx)
	assert.Equal(t, 2, count)

	// get
	_, err = s.Get(ctx, 1)
	assert.NotNil(t, err)
	message, err = s.Get(ctx, id)
	assert.Nil(t, err)
	assert.Equal(t, "test updated", message.Body)
	assert.Equal(t, id, message.ID)

	// query
	messages, _ := s.Query(ctx, "connection", 0, 0, 0)
	assert.Equal(t, 2, len(messages))

	// delete
	_, err = s.Delete(ctx, 1)
	assert.NotNil(t, err)
	message, err = s.Delete(ctx, id)
	assert.Nil(t, err)
	assert.Equal(t, id, message.ID)
	count, _ = s.Count(ctx)
	assert.Equal(t, 1, count)
}
