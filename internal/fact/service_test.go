package fact

import (
	"context"
	"testing"

	"github.com/joinself/restful-client/pkg/log"
	"github.com/joinself/restful-client/pkg/mock"
	"github.com/stretchr/testify/assert"
)

func TestCreateFactRequest_Validate(t *testing.T) {
	tests := []struct {
		name      string
		model     CreateFactRequest
		wantError bool
	}{
		{"success", CreateFactRequest{
			Facts: []FactToIssue{{
				Key:    "test",
				Value:  "test",
				Source: "test",
			}},
		}, false},
		{"required", CreateFactRequest{
			Facts: []FactToIssue{{
				Key: "test",
			}},
		}, true},
		{"required-key", CreateFactRequest{
			Facts: []FactToIssue{{
				Value: "test",
			}},
		}, true},
		{"too long", CreateFactRequest{
			Facts: []FactToIssue{{
				Key:    "1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890",
				Value:  "test",
				Source: "test",
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

func Test_service_CRUD(t *testing.T) {
	logger, _ := log.NewForTest()
	runner := mock.NewRunnerMock()
	s := NewService(&mock.FactRepositoryMock{}, &mock.AttestationRepositoryMock{}, runner, logger)

	ctx := context.Background()

	// initial count
	count, _ := s.Count(ctx, 1, "", "")
	assert.Equal(t, 0, count)

	// successful creation
	err := s.Create(ctx, "app", "connection", 1, CreateFactRequest{
		Facts: []FactToIssue{{
			Key:   "test",
			Value: "test",
		}},
	})
	assert.Nil(t, err)
}
