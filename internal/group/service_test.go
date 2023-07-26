package group

import (
	"context"
	"errors"
	"testing"

	"github.com/joinself/restful-client/internal/entity"
	"github.com/joinself/restful-client/internal/test"
	"github.com/joinself/restful-client/pkg/log"
	"github.com/joinself/restful-client/pkg/mock"
	"github.com/stretchr/testify/assert"
)

var errCRUD = errors.New("error crud")

func TestCreateGroupRequest_Validate(t *testing.T) {
	tests := []struct {
		name      string
		model     CreateGroupRequest
		wantError bool
	}{
		{"success", CreateGroupRequest{Name: "selfid", Members: []string{"1112223334", "1112223335"}}, false},
		{"no members", CreateGroupRequest{Name: "selfid"}, true},
		{"no name", CreateGroupRequest{Members: []string{"1112223334", "1112223335"}}, true},
		{"empty members", CreateGroupRequest{Name: "selfid", Members: []string{}}, true},
		{"required", CreateGroupRequest{}, true},
		{"too long", CreateGroupRequest{Name: "1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.model.Validate()
			assert.Equal(t, tt.wantError, err != nil)
		})
	}
}

func TestUpdateGroupRequest_Validate(t *testing.T) {
	tests := []struct {
		name      string
		model     UpdateGroupRequest
		wantError bool
	}{
		{"success", UpdateGroupRequest{Name: "test"}, false},
		{"required", UpdateGroupRequest{Name: ""}, true},
		{"too long", UpdateGroupRequest{Name: "1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.model.Validate()
			assert.Equal(t, tt.wantError, err != nil)
		})
	}
}

func Test_service_CRUD(t *testing.T) {
	appID := "appID"
	ctx := context.Background()
	logger, _ := log.NewForTest()

	db := test.DB(t)
	test.ResetAll(t, db)

	cMock := mock.ConnectionRepositoryMock{}
	cMock.Create(ctx, entity.Connection{
		AppID:  appID,
		SelfID: "1112223334",
	})
	cMock.Create(ctx, entity.Connection{
		AppID:  appID,
		SelfID: "1112223335",
	})
	cMock.Create(ctx, entity.Connection{
		AppID:  appID,
		SelfID: "1112223336",
	})

	s := NewService(&mock.GroupRepositoryMock{}, &cMock, logger, nil)

	// initial count
	count, _ := s.Count(ctx)
	assert.Equal(t, 0, count)

	// successful creation
	group, err := s.Create(ctx, appID, CreateGroupRequest{
		Name:    "test",
		Members: []string{"1112223334", "1112223335"},
	})
	assert.Nil(t, err)
	assert.Equal(t, "test", group.Name)
	assert.Equal(t, 2, len(group.Members))
	assert.NotEmpty(t, group.CreatedAt)
	assert.NotEmpty(t, group.UpdatedAt)
	count, _ = s.Count(ctx)
	assert.Equal(t, 1, count)

	group, err = s.Get(ctx, appID, group.ID)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(group.Members))
	// validation error in creation

	group, err = s.Create(ctx, appID, CreateGroupRequest{
		Name:    "",
		Members: []string{"1112223334", "1112223335"},
	})

	assert.NotNil(t, err)
	count, _ = s.Count(ctx)
	assert.Equal(t, 1, count)

	// unexpected error in creation

	group, err = s.Create(ctx, appID, CreateGroupRequest{
		Name:    "error",
		Members: []string{"1112223334", "1112223335"},
	})

	assert.Equal(t, errCRUD, err)
	count, _ = s.Count(ctx)
	assert.Equal(t, 1, count)

	group, err = s.Create(ctx, appID, CreateGroupRequest{
		Name:    "test2r",
		Members: []string{"1112223334", "1112223335"},
	})

	// update
	group, err = s.Update(ctx, appID, group.ID, UpdateGroupRequest{
		Name: "test updated",
		Members: []string{
			"1112223334",
			"1112223335",
			"1112223336",
		},
	})

	assert.Nil(t, err)
	assert.Equal(t, "test updated", group.Name)

	group, err = s.Get(ctx, appID, group.ID)
	assert.Nil(t, err)
	assert.Equal(t, "test updated", group.Name)
	assert.Equal(t, 3, len(group.Members))

	// update with an unexisting connection!!
	group, err = s.Update(ctx, appID, group.ID, UpdateGroupRequest{
		Name: "test updated",
		Members: []string{
			"1112223334",
			"1112223335",
			"1112223336",
			"1000000000",
		},
	})

	assert.Nil(t, err)
	assert.Equal(t, "test updated", group.Name)

	group, err = s.Get(ctx, appID, group.ID)
	assert.Nil(t, err)
	assert.Equal(t, "test updated", group.Name)
	assert.Equal(t, 3, len(group.Members))

	// validation error in update
	_, err = s.Update(ctx, appID, group.ID, UpdateGroupRequest{
		Name: "",
	})

	assert.NotNil(t, err)
	count, _ = s.Count(ctx)
	assert.Equal(t, 2, count)

	// get
	_, err = s.Get(ctx, appID, 9000)
	assert.NotNil(t, err)
	retrievedGroup, err := s.Get(ctx, appID, group.ID)
	assert.Nil(t, err)
	assert.Equal(t, "test updated", group.Name)
	assert.Equal(t, group.ID, retrievedGroup.ID)

	// query
	groups, _ := s.Query(ctx, appID, 0, 0)
	assert.Equal(t, 2, len(groups))

	// delete
	err = s.Delete(ctx, appID, 1)
	assert.NotNil(t, err)
	err = s.Delete(ctx, appID, group.ID)
	assert.Nil(t, err)
	count, _ = s.Count(ctx)
	assert.Equal(t, 1, count)

}
