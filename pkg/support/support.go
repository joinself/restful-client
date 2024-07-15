package support

import (
	"github.com/joinself/restful-client/pkg/webhook"
	selfsdk "github.com/joinself/self-go-sdk"
	"github.com/joinself/self-go-sdk/messaging"
)

type SelfClient interface {
	Start() error
	SelfAppID() string
	MessagingService() MessagingService
	ChatService() ChatService
	FactService() FactService
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
func (s *selfClient) FactService() FactService {
	return s.client.FactService()
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
type FactService interface{}

type SelfClientGetter interface {
	Get(id string) (*selfsdk.Client, bool)
	Poster(id string) (webhook.Poster, bool)
}
