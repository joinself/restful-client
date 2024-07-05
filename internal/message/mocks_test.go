package message

import (
	"context"
	"errors"

	"github.com/joinself/restful-client/internal/connection"
	"github.com/joinself/restful-client/internal/entity"
)

type mockService struct{}

func (m mockService) Get(ctx context.Context, connectionID int, jti string) (Message, error) {
	if jti == "not_found_id" {
		return Message{}, errors.New("not found")
	}

	return Message{
		Body:         "body",
		ConnectionID: "iss",
		CID:          "cid",
	}, nil
}

func (m mockService) Query(ctx context.Context, connection int, messagesSince int, offset, limit int) ([]Message, error) {
	if messagesSince == 98 {
		return []Message{}, errors.New("expected error")
	}
	return []Message{}, nil
}
func (m mockService) Count(ctx context.Context, connectionID, messagesSince int) (int, error) {
	if messagesSince == 99 {
		return 0, errors.New("expected count error")
	}
	return 0, nil
}
func (m mockService) Create(ctx context.Context, appID, connectionID string, connection int, input CreateMessageRequest) (Message, error) {
	if input.Body == "error" {
		return Message{}, errors.New("error!")
	}
	return Message{}, nil
}
func (m mockService) Update(ctx context.Context, appID string, connectionID int, selfID string, jti string, req UpdateMessageRequest) (Message, error) {
	if req.Body == "error" {
		return Message{}, errors.New("error!")
	}
	return Message{}, nil
}
func (m mockService) Delete(ctx context.Context, connectionID int, jti string) error {
	if jti == "error" {
		return errors.New("error!")
	}
	return nil
}
func (m mockService) MarkAsRead(ctx context.Context, appID, connection, jti string, connectionID int) error {
	return nil
}
func (m mockService) MarkAsReceived(ctx context.Context, appID, connection, jti string, connectionID int) error {
	return nil
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
