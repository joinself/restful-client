package mock

import (
	"github.com/joinself/restful-client/internal/entity"
	"github.com/joinself/restful-client/pkg/webhook"
	selfsdk "github.com/joinself/self-go-sdk"
)

type RunnerMock struct {
	apps map[string]bool
}

func NewRunnerMock() *RunnerMock {
	return &RunnerMock{
		apps: map[string]bool{},
	}
}

func (m RunnerMock) Run(app entity.App) error {
	m.apps[app.ID] = true
	return nil
}

func (m RunnerMock) Stop(id string) {
	m.apps[id] = false
}

func (m RunnerMock) StopAll() {
	for id, _ := range m.apps {
		m.apps[id] = false
	}
}

func (m RunnerMock) Get(id string) (*selfsdk.Client, bool) {
	return nil, false
}

func (m RunnerMock) Poster(id string) (webhook.Poster, bool) {
	return nil, false
}

func (m RunnerMock) SetApp(app entity.App) error {
	return nil
}
