package mock

import (
	"github.com/joinself/restful-client/pkg/webhook"
)

type PosterMock struct {
	History []webhook.WebhookPayload
}

func (p *PosterMock) Post(payload webhook.WebhookPayload) error {
	p.History = append(p.History, payload)
	return nil
}
