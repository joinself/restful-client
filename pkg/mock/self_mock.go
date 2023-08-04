package mock

import (
	"github.com/joinself/restful-client/internal/entity"
	"github.com/joinself/restful-client/pkg/support"
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

type SelfMessenger struct {
}

func (s *SelfMessenger) Subscribe(messageType string, h func(m *messaging.Message)) {

}
