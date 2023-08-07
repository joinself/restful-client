package request

import (
	"context"
	"testing"

	"github.com/joinself/restful-client/pkg/log"
	"github.com/joinself/restful-client/pkg/mock"
	"github.com/joinself/restful-client/pkg/webhook"
	"github.com/stretchr/testify/assert"
)

func TestCreateMessageRequest_Validate(t *testing.T) {
	tests := []struct {
		name      string
		model     CreateRequest
		wantError bool
	}{
		{"auth", CreateRequest{Type: "auth"}, false},
		{"fact", CreateRequest{Type: "fact"}, false},
		{"invalid", CreateRequest{Type: "invalid"}, true},
		{"empty", CreateRequest{Type: ""}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.model.Validate()
			assert.Equal(t, tt.wantError, err != nil)
		})
	}
}

func TestService_CRUD(t *testing.T) {
	logger, _ := log.NewForTest()
	m := mock.RequestRepositoryMock{}
	s := NewService(
		&m,
		&mock.FactRepositoryMock{},
		&mock.AttestationRepositoryMock{},
		logger,
		map[string]RequesterService{},
		map[string]*webhook.Webhook{},
	)
	ctx := context.Background()

	connection := 1

	_, err := s.Create(ctx, "app", "connection", connection, CreateRequest{Type: ""})
	assert.Error(t, err)

	assert.Equal(t, 0, len(m.Items))
	req, err := s.Create(ctx, "app", "connection", connection, CreateRequest{Type: "auth"})
	assert.NoError(t, err)
	assert.Equal(t, "auth", req.Type)
	assert.Equal(t, 1, len(m.Items))

	qReq, err := s.Get(ctx, "app", "connection", req.ID)
	assert.NoError(t, err)
	assert.Equal(t, req.ID, qReq.ID)

	_, err = s.Get(ctx, "app", "connection", "unexisting")
	assert.Error(t, err)

}
