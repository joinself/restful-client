package request

import (
	"context"
	"errors"

	"github.com/joinself/restful-client/internal/connection"
	"github.com/joinself/restful-client/internal/entity"
	"github.com/joinself/restful-client/pkg/support"
	selffact "github.com/joinself/self-go-sdk/fact"
)

type mockService struct{}

func (m mockService) Get(ctx context.Context, appID, id string) (ExtRequest, error) {
	if id == "not_found_id" {
		return ExtRequest{}, errors.New("not found")
	}

	return ExtRequest{}, nil
}

func (m mockService) Create(ctx context.Context, appID string, conn *entity.Connection, input CreateRequest) (ExtRequest, error) {
	if appID == "error" {
		return ExtRequest{}, errors.New("error!")
	}
	return ExtRequest{}, nil
}

func (m mockService) CreateFactsFromResponse(conn entity.Connection, req entity.Request, facts []selffact.Fact) []entity.Fact {
	return []entity.Fact{}
}

func (m mockService) SetRunner(runner support.SelfClientGetter) {
	return
}

type mockConnectionService struct{}

func (m mockConnectionService) Get(ctx context.Context, appid, selfid string) (connection.Connection, error) {
	if selfid == "not_found_id" {
		return connection.Connection{}, errors.New("expected not found")
	}
	return connection.Connection{
		entity.Connection{
			SelfID: "selfid",
			AppID:  appid,
			Name:   "name",
		},
	}, nil
}

func (m mockConnectionService) Query(ctx context.Context, appid string, offset, limit int) ([]connection.Connection, error) {
	return []connection.Connection{}, nil
}

func (m mockConnectionService) Count(ctx context.Context, appid string) (int, error) {
	if appid == "count_error" {
		return 0, errors.New("expected_error_count")
	}
	return 1, nil
}

func (m mockConnectionService) Create(ctx context.Context, appid string, input connection.CreateConnectionRequest) (connection.Connection, error) {
	if input.SelfID == "controlled_error" {
		return connection.Connection{}, errors.New("controlled error")
	}
	return connection.Connection{}, nil
}

func (m mockConnectionService) Update(ctx context.Context, appid, selfid string, input connection.UpdateConnectionRequest) (connection.Connection, error) {
	if input.Name == "controlled_error" {
		return connection.Connection{}, errors.New("controlled error")
	}
	return connection.Connection{}, nil
}

func (m mockConnectionService) Delete(ctx context.Context, appid, selfid string) (connection.Connection, error) {
	if selfid == "controlled_error" {
		return connection.Connection{}, errors.New("controlled error")
	}
	return connection.Connection{}, nil
}
