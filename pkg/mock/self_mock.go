package mock

import (
	"github.com/joinself/restful-client/internal/entity"
	"github.com/joinself/restful-client/pkg/support"
	selfsdk "github.com/joinself/self-go-sdk"
	"github.com/joinself/self-go-sdk/fact"
	"github.com/joinself/self-go-sdk/messaging"
)

type SelfMock struct {
	Started bool
	Items   []entity.Message
}

func (m *SelfMock) Start() error {
	m.Started = true
	return nil
}

func (m *SelfMock) SelfAppID() string {
	return "test"
}

func (m *SelfMock) MessagingService() support.MessagingService {
	return &SelfMessenger{}
}

func (m *SelfMock) ChatService() support.ChatService {
	return nil
}

func (m *SelfMock) FactService() support.FactService {
	return &FactServiceMock{}
}

func (s *SelfMock) Stop() {
}

func (s *SelfMock) Get() *selfsdk.Client {
	return nil
}

type FactServiceMock struct {
}

func (s *FactServiceMock) FactResponse(issuer, subject string, response []byte) ([]fact.Fact, error) {
	facts := make([]fact.Fact, 1)
	facts[0] = fact.Fact{
		Fact: fact.FactDisplayName,
	}
	return facts, nil
}

func (s *FactServiceMock) RequestAsync(req *fact.FactRequestAsync) error {
	return nil
}

type SelfMessenger struct {
}

func (s *SelfMessenger) Subscribe(messageType string, h func(m *messaging.Message)) {

}
