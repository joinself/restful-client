package fact

import (
	"context"
	"testing"

	"github.com/joinself/restful-client/pkg/log"
	"github.com/joinself/restful-client/pkg/mock"
	"github.com/joinself/self-go-sdk/fact"
	"github.com/stretchr/testify/assert"
)

func TestCreateFactRequest_Validate(t *testing.T) {
	tests := []struct {
		name      string
		model     CreateFactRequest
		wantError bool
	}{
		{"success", CreateFactRequest{
			Facts: []fact.FactToIssue{{
				Key:   "test",
				Value: "test",
			}},
		}, false},
		{"required", CreateFactRequest{
			Facts: []fact.FactToIssue{{
				Key: "test",
			}},
		}, true},
		{"required-key", CreateFactRequest{
			Facts: []fact.FactToIssue{{
				Value: "test",
			}},
		}, true},
		{"too long", CreateFactRequest{
			Facts: []fact.FactToIssue{{
				Key:   "1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890",
				Value: "test",
			}},
		}, true},
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
	err := s.Create(ctx, "app", "connection", 1, CreateFactRequest{
		Facts: []fact.FactToIssue{{
			Key:   "test",
			Value: "test",
		}},
	})
	assert.Nil(t, err)

	// validation error in creation
	err = s.Create(ctx, "app", "connection", 1, CreateFactRequest{
		Facts: []fact.FactToIssue{{
			Value: "test",
		}},
	})
	assert.NotNil(t, err)

	// unexpected error in creation
	err = s.Create(ctx, "app", "connection", 1, CreateFactRequest{
		Facts: []fact.FactToIssue{{
			Key:   "",
			Value: "test",
		}},
	})
	assert.Error(t, err)
}
