package mock

import (
	"github.com/joinself/restful-client/pkg/webhook"
	"github.com/joinself/restful-client/pkg/worker"
)

type CallbackWorkerPoolMock struct {
	History []webhook.WebhookPayload
}

func (p *CallbackWorkerPoolMock) Send(qm worker.CallbackTask) error {
	p.History = append(p.History, qm.WebhookPayload)
	return nil
}
