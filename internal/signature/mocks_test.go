package signature

import (
	"context"
	"errors"

	"github.com/joinself/restful-client/internal/connection"
	"github.com/joinself/restful-client/internal/entity"
)

type mockService struct{}

func (m mockService) Get(ctx context.Context, aID, cID, id string) (ExtSignature, error) {
	if id == "not_found_id" {
		return ExtSignature{}, errors.New("not found")
	}

	return ExtSignature{
		Description: "body",
	}, nil
}

func (m mockService) Query(ctx context.Context, aID, cID string, signaturesSince int, offset, limit int) ([]ExtSignature, error) {
	if signaturesSince == 98 {
		return []ExtSignature{}, errors.New("expected error")
	}
	return []ExtSignature{}, nil
}
func (m mockService) Count(ctx context.Context, aID, cID string, signaturesSince int) (int, error) {
	if signaturesSince == 99 {
		return 0, errors.New("expected count error")
	}
	return 0, nil
}
func (m mockService) Create(ctx context.Context, appID, connectionID string, input CreateSignatureRequest) (ExtSignature, error) {
	if input.Description == "error" {
		return ExtSignature{}, errors.New("error!")
	}
	return ExtSignature{}, nil
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
	conns := []connection.Connection{}
	if appid == "query_error" {
		return []connection.Connection{}, errors.New("expected_error_query")
	}
	conns = append(conns, connection.Connection{
		entity.Connection{
			SelfID: "selfid",
			AppID:  appid,
			Name:   "name",
		},
	})
	return conns, nil
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
