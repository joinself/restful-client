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
	s := NewService(&mock.VoiceRepositoryMock{}, runner, logger)
	ctx := context.Background()

	appID, _ := uuid.NewV4()
	connection, _ := uuid.NewV4()
	callID, _ := uuid.NewV4()

	// successful creation
	call, err := s.Setup(ctx, appID.String(), connection.String(), callID.String())
	assert.Nil(t, err)
	assert.Equal(t, callID.String(), call.CallID)

	// update
	err = s.Start(ctx, appID.String(), connection.String(), callID.String(), ProceedData{PeerInfo: "pper_info"})
	assert.Nil(t, err)
	assert.Equal(t, "started", call.Status)

	// get
	call, err = s.Get(ctx, appID.String(), connection.String(), callID.String())
	assert.NotNil(t, err)
	assert.Equal(t, "started", call.Status)
}
*/
