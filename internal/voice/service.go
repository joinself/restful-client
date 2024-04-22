package voice

import (
	"context"
	"errors"
	"log"

	"github.com/joinself/restful-client/pkg/support"
)

type Service interface {
	Start(ctx context.Context, appID, connectionID, callID string, data ProceedData) error
	Accept(ctx context.Context, appID, recipient, callID string, data ProceedData) error
	Busy(ctx context.Context, appID, recipient, callID, cid string) error
	Stop(ctx context.Context, appID, recipient, callID, cid string) error
	Setup(ctx context.Context, appID, recipient, name, callID string) error
}

type service struct {
	runner support.SelfClientGetter
	logger log.Logger
}

func NewService(runner support.SelfClientGetter, logger log.Logger) Service {
	return service{runner, logger}
}

func (s service) Setup(ctx context.Context, appID, recipient, name, callID string) error {
	c, ok := s.runner.Get(appID)
	if !ok {
		return errors.New("app not configured or started")
	}

	return c.VoiceService().Setup(recipient, name, callID)
}

func (s service) Start(ctx context.Context, appID, recipient, callID string, data ProceedData) error {
	c, ok := s.runner.Get(appID)
	if !ok {
		return errors.New("app not configured or started")
	}

	return c.VoiceService().Start(recipient, data.CID, callID, data.PeerInfo, data.Data)
}

func (s service) Accept(ctx context.Context, appID, recipient, callID string, data ProceedData) error {
	c, ok := s.runner.Get(appID)
	if !ok {
		return errors.New("app not configured or started")
	}

	return c.VoiceService().Accept(recipient, data.CID, callID, data.PeerInfo, data.Data)
}

func (s service) Busy(ctx context.Context, appID, recipient, callID, cid string) error {
	c, ok := s.runner.Get(appID)
	if !ok {
		return errors.New("app not configured or started")
	}

	return c.VoiceService().Stop(recipient, cid, callID)
}

func (s service) Stop(ctx context.Context, appID, recipient, callID, cid string) error {
	c, ok := s.runner.Get(appID)
	if !ok {
		return errors.New("app not configured or started")
	}

	return c.VoiceService().Stop(recipient, cid, callID)
}
