package voice

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

var errCRUD = errors.New("error crud")

const LARGE_STRING = "1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890"

func TestSetupRequest_Validate(t *testing.T) {
	tests := []struct {
		name      string
		model     SetupData
		wantError bool
	}{
		{"success", SetupData{Name: "test"}, false},
		{"required", SetupData{Name: ""}, true},
		{"too long", SetupData{Name: LARGE_STRING}, true},
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
		model     ProceedData
		wantError bool
	}{
		{"success", ProceedData{PeerInfo: "test"}, false},
		{"required", ProceedData{PeerInfo: ""}, true},
		{"too long", ProceedData{PeerInfo: LARGE_STRING}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.model.Validate()
			assert.Equal(t, tt.wantError, err != nil)
		})
	}
}

/*
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
	voice, err := s.Create(ctx, "app", "connection", connection, CreateMessageRequest{Body: "test"})
	assert.Nil(t, err)
	id := voice.ID
	assert.Equal(t, "test", voice.Body)
	assert.NotEmpty(t, voice.CreatedAt)
	assert.NotEmpty(t, voice.UpdatedAt)
	count, _ = s.Count(ctx, connection, 0)
	assert.Equal(t, 1, count)

	_, _ = s.Create(ctx, "app", "connection", connection, CreateMessageRequest{Body: "test2"})

	// update
	voice, err = s.Update(ctx, "app", connection, "connection", voice.ID, UpdateMessageRequest{Body: "test updated"})
	assert.Nil(t, err)
	assert.Equal(t, "test updated", voice.Body)
	_, err = s.Update(ctx, "app", connection, "connection", "1", UpdateMessageRequest{Body: "test updated"})
	assert.NotNil(t, err)

	// get
	_, err = s.Get(ctx, connection, "1")
	assert.NotNil(t, err)
	voice, err = s.Get(ctx, connection, voice.ID)
	assert.Nil(t, err)
	assert.Equal(t, "test updated", voice.Body)
	assert.Equal(t, id, voice.ID)

	// query
	voices, _ := s.Query(ctx, connection, 0, 0, 0)
	assert.Equal(t, 2, len(voices))

	// delete
	err = s.Delete(ctx, connection, "non existing")
	assert.NotNil(t, err)
	err = s.Delete(ctx, connection, voice.ID)
	assert.Nil(t, err)
	count, _ = s.Count(ctx, connection, 0)
	assert.Equal(t, 1, count)
}
*/
