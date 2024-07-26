package worker

import (
	"github.com/joinself/restful-client/pkg/webhook"
)

// CallbackTask is the task to be queued.
type CallbackTask struct {
	AppID          string                 `json:"app_id"`
	WebhookPayload webhook.WebhookPayload `json:"webhook"`
}

// Send executes the send operation, so a webhook is sent to
// the configured callback url.
func (ct *CallbackTask) Send(s CallbackSender) error {
	return s.SendCallback(ct.AppID, ct.WebhookPayload)
}
