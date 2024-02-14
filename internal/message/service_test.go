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
	runner := mock.NewRunnerMock()
	s := NewService(&mock.MessageRepositoryMock{}, runner, logger)
	ctx := context.Background()

	connection := 1

	// initial count
	count, _ := s.Count(ctx, connection, 0)
	assert.Equal(t, 0, count)

	// successful creation
	message, err := s.Create(ctx, "app", "connection", connection, CreateMessageRequest{Body: "test"})
	assert.Nil(t, err)
	id := message.ID
	assert.Equal(t, "test", message.Body)
	assert.NotEmpty(t, message.CreatedAt)
	assert.NotEmpty(t, message.UpdatedAt)
	count, _ = s.Count(ctx, connection, 0)
	assert.Equal(t, 1, count)

	_, _ = s.Create(ctx, "app", "connection", connection, CreateMessageRequest{Body: "test2"})

	// update
	message, err = s.Update(ctx, "app", connection, "connection", message.JTI, UpdateMessageRequest{Body: "test updated"})
	assert.Nil(t, err)
	assert.Equal(t, "test updated", message.Body)
	_, err = s.Update(ctx, "app", connection, "connection", "1", UpdateMessageRequest{Body: "test updated"})
	assert.NotNil(t, err)

	// get
	_, err = s.Get(ctx, connection, "1")
	assert.NotNil(t, err)
	message, err = s.Get(ctx, connection, message.JTI)
	assert.Nil(t, err)
	assert.Equal(t, "test updated", message.Body)
	assert.Equal(t, id, message.ID)

	// query
	messages, _ := s.Query(ctx, connection, 0, 0, 0)
	assert.Equal(t, 2, len(messages))

	// delete
	err = s.Delete(ctx, connection, "non existing")
	assert.NotNil(t, err)
	err = s.Delete(ctx, connection, message.JTI)
	assert.Nil(t, err)
	count, _ = s.Count(ctx, connection, 0)
	assert.Equal(t, 1, count)
}
