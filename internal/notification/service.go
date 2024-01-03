package notification

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/joinself/restful-client/pkg/log"
	selfsdk "github.com/joinself/self-go-sdk"
)

// Service encapsulates usecase logic for messages.
type Service interface {
	Send(ctx context.Context, appID, selfID string, notification SystemNotificationData) error
}

// SystemNotification
type SystemNotificationData struct {
	Notification *struct {
		Type    string `json:"type"`
		Title   string `json:"title"`
		Message string `json:"message"`
	} `json:"notification,omitempty"`
	Metadata *struct {
		Type    string `json:"type"`
		Payload string `json:"payload"`
	} `json:"metadata,omitempty"`
}

// Validate validates the SystemNotification fields.
func (m SystemNotificationData) Validate() error {
	return nil
}

type SelfClientGetter interface {
	Get(id string) (*selfsdk.Client, bool)
}

type service struct {
	logger log.Logger
	runner SelfClientGetter
}

// NewService creates a new notification service.
func NewService(runner SelfClientGetter, logger log.Logger) Service {
	return service{logger, runner}
}

// Send sends a system notification.
func (s service) Send(ctx context.Context, appID, selfID string, data SystemNotificationData) error {
	client, ok := s.runner.Get(appID)
	if !ok {
		return nil
	}

	cid := uuid.New().String()
	req, err := client.MessagingService().BuildRequest(map[string]interface{}{
		"typ":  "system.notify",
		"sub":  selfID,
		"cid":  cid,
		"data": data,
	})
	if err != nil {
		return err
	}

	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	devices, err := client.IdentityService().GetDevices(selfID)
	if err != nil {
		return err
	}

	recipients := []string{}
	for _, device := range devices {
		recipients = append(recipients, selfID+":"+string(device))
	}
	if len(recipients) == 0 {
		return nil
	}

	err = client.MessagingService().Send(recipients, cid, body)
	if err != nil {
		return err
	}

	return nil
}
