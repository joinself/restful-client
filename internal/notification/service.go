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

type service struct {
	logger  log.Logger
	clients map[string]*selfsdk.Client
}

// NewService creates a new notification service.
func NewService(logger log.Logger, clients map[string]*selfsdk.Client) Service {
	return service{logger, clients}
}

// Send sends a system notification.
func (s service) Send(ctx context.Context, appID, selfID string, data SystemNotificationData) error {
	if _, ok := s.clients[appID]; !ok {
		return nil
	}

	cid := uuid.New().String()
	req, err := s.clients[appID].MessagingService().BuildRequest(map[string]interface{}{
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

	devices, err := s.clients[appID].IdentityService().GetDevices(selfID)
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

	err = s.clients[appID].MessagingService().Send(recipients, cid, body)
	if err != nil {
		return err
	}

	return nil
}
