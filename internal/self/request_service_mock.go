package self

import (
	"context"
	"database/sql"

	"github.com/joinself/restful-client/internal/entity"
	"github.com/joinself/restful-client/internal/request"
	selffact "github.com/joinself/self-go-sdk/fact"
)

type RequestServiceMock struct {
	Items []request.Request
}

func (m RequestServiceMock) Get(ctx context.Context, appID, id string) (request.Request, error) {
	for _, item := range m.Items {
		if item.ID == id {
			return item, nil
		}
	}
	return request.Request{}, sql.ErrNoRows
}

func (m *RequestServiceMock) Create(ctx context.Context, appID string, connection *entity.Connection, input request.CreateRequest) (request.Request, error) {
	r := request.Request{}
	m.Items = append(m.Items, r)
	return r, nil
}

func (m *RequestServiceMock) CreateFactsFromResponse(conn entity.Connection, req entity.Request, facts []selffact.Fact) []entity.Fact {
	r := request.Request{}
	m.Items = append(m.Items, r)
	return nil
}
