package signature

import (
	"context"
	"errors"
	"testing"

	"github.com/joinself/restful-client/pkg/log"
	"github.com/joinself/restful-client/pkg/mock"
	"github.com/stretchr/testify/assert"
)

var errCRUD = errors.New("error crud")

const LARGE_STRING = "1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890"

func TestCreateRequest_Validate(t *testing.T) {
	tests := []struct {
		name      string
		model     CreateSignatureRequest
		wantError bool
	}{
		{"success", CreateSignatureRequest{Description: "test"}, false},
		{"required", CreateSignatureRequest{Description: ""}, true},
		{"too long", CreateSignatureRequest{Description: LARGE_STRING}, true},
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
	s := NewService(&mock.SignatureRepositoryMock{}, runner, logger)
	ctx := context.Background()

	app := "app"
	connection := "conection"

	// initial count
	count, _ := s.Count(ctx, app, connection, 0)
	assert.Equal(t, 0, count)

	// successful creation
	signature, err := s.Create(ctx, app, connection, CreateSignatureRequest{Description: "test"})
	assert.Nil(t, err)
	assert.Equal(t, "test", signature.Description)
	assert.NotEmpty(t, signature.CreatedAt)
	assert.NotEmpty(t, signature.UpdatedAt)
	count, _ = s.Count(ctx, app, connection, 0)
	assert.Equal(t, 1, count)

	_, _ = s.Create(ctx, app, connection, CreateSignatureRequest{Description: "test2"})

	// query
	signatures, _ := s.Query(ctx, app, connection, 0, 0, 100)
	assert.Equal(t, 2, len(signatures))
}
