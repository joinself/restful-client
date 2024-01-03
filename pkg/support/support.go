package support

import (
	selfsdk "github.com/joinself/self-go-sdk"
	"github.com/joinself/self-go-sdk/messaging"
)

type SelfClient interface {
	Start() error
	SelfAppID() string
	MessagingService() MessagingService
	ChatService() ChatService
	Stop()
	Get() *selfsdk.Client
}

type selfClient struct {
	client *selfsdk.Client
}

func (s *selfClient) Start() error {
	return s.client.Start()
}
func (s *selfClient) SelfAppID() string {
	return s.client.SelfAppID()
}
func (s *selfClient) MessagingService() MessagingService {
	return s.client.MessagingService()
}
func (s *selfClient) ChatService() ChatService {
	return s.client.ChatService()
}
func (s *selfClient) Stop() {
	s.client.Close()
}
func (s *selfClient) Get() *selfsdk.Client {
	return s.client
}

func NewSelfClient(client *selfsdk.Client) SelfClient {
	return &selfClient{client}
}

type MessagingService interface {
	Subscribe(messageType string, h func(m *messaging.Message))
}

type ChatService interface{}
