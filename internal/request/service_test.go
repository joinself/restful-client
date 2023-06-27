package request

import (
	"testing"

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
