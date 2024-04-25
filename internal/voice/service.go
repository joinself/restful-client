package voice

import (
	"context"
	"errors"

	"github.com/gofrs/uuid"
	"github.com/joinself/restful-client/internal/entity"
	"github.com/joinself/restful-client/pkg/log"
	"github.com/joinself/restful-client/pkg/support"
)

type Service interface {
	Get(ctx context.Context, appID, recipient, callID string) (ExtCall, error)
	Start(ctx context.Context, appID, connectionID, callID string, data ProceedData) error
	Accept(ctx context.Context, appID, recipient, callID string, data ProceedData) error
	Busy(ctx context.Context, appID, recipient, callID string) error
	Stop(ctx context.Context, appID, recipient, callID string) error
	Setup(ctx context.Context, appID, recipient, name string) (*entity.Call, error)
	// Count returns the number of calls.
	Count(ctx context.Context, aID, cID string, callsSince int) (int, error)
	// Query returns the calls with the specified offset and limit.
	Query(ctx context.Context, aID, cID string, callsSince int, offset, limit int) ([]ExtCall, error)
}

type service struct {
	repo   Repository
	runner support.SelfClientGetter
	logger log.Logger
}

func NewService(repo Repository, runner support.SelfClientGetter, logger log.Logger) Service {
	return service{repo, runner, logger}
}

func (s service) Get(ctx context.Context, appID, recipient, callID string) (ExtCall, error) {
	call, err := s.repo.Get(ctx, appID, recipient, callID)
	return newExtCall(call), err
}

func (s service) Setup(ctx context.Context, appID, recipient, name string) (*entity.Call, error) {
	c, ok := s.runner.Get(appID)
	if !ok {
		return nil, errors.New("app not configured or started")
	}

	callID, err := uuid.NewV4()
	if err != nil {
		return nil, errors.New("error generating uuid")
	}

	call := entity.Call{
		AppID:  appID,
		SelfID: recipient,
		CallID: callID.String(),
	}

	err = s.repo.Create(ctx, &call)
	if err != nil {
		return nil, errors.New("error creating the app : " + err.Error())
	}

	return &call, c.VoiceService().Setup(recipient, name, callID.String())
}

func (s service) Start(ctx context.Context, appID, recipient, callID string, data ProceedData) error {
	c, ok := s.runner.Get(appID)
	if !ok {
		return errors.New("app not configured or started")
	}

	call, err := s.repo.Get(ctx, appID, recipient, callID)
	if err != nil {
		return errors.New("app does not exist : " + err.Error())
	}

	call.Status = "started"
	call.PeerInfo = data.PeerInfo

	err = s.repo.Update(ctx, call)
	if err != nil {
		return errors.New("error updating the call : " + err.Error())
	}

	cid, err := uuid.NewV4()
	if err != nil {
		return errors.New("error generating cid")
	}

	return c.VoiceService().Start(recipient, cid.String(), callID, data.PeerInfo, map[string]interface{}{
		"name": data.Name,
	})
}

func (s service) Accept(ctx context.Context, appID, recipient, callID string, data ProceedData) error {
	c, ok := s.runner.Get(appID)
	if !ok {
		return errors.New("app not configured or started")
	}

	call, err := s.repo.Get(ctx, appID, recipient, callID)
	if err != nil {
		return errors.New("app does not exist : " + err.Error())
	}

	call.Status = "accepted"
	call.PeerInfo = data.PeerInfo

	err = s.repo.Update(ctx, call)
	if err != nil {
		return errors.New("error updating the call : " + err.Error())
	}

	cid, err := uuid.NewV4()
	if err != nil {
		return errors.New("error generating cid")
	}

	return c.VoiceService().Accept(recipient, cid.String(), callID, data.PeerInfo, map[string]interface{}{
		"name": data.Name,
	})
}

func (s service) Busy(ctx context.Context, appID, recipient, callID string) error {
	c, ok := s.runner.Get(appID)
	if !ok {
		return errors.New("app not configured or started")
	}

	call, err := s.repo.Get(ctx, appID, recipient, callID)
	if err != nil {
		return errors.New("app does not exist : " + err.Error())
	}

	call.Status = "busy"

	err = s.repo.Update(ctx, call)
	if err != nil {
		return errors.New("error updating the call : " + err.Error())
	}

	cid, err := uuid.NewV4()
	if err != nil {
		return errors.New("error generating cid")
	}

	return c.VoiceService().Stop(recipient, cid.String(), callID)
}

func (s service) Stop(ctx context.Context, appID, recipient, callID string) error {
	c, ok := s.runner.Get(appID)
	if !ok {
		return errors.New("app not configured or started")
	}

	call, err := s.repo.Get(ctx, appID, recipient, callID)
	if err != nil {
		return errors.New("app does not exist : " + err.Error())
	}

	call.Status = "ended"

	err = s.repo.Update(ctx, call)
	if err != nil {
		return errors.New("error updating the call : " + err.Error())
	}

	cid, err := uuid.NewV4()
	if err != nil {
		return errors.New("error generating cid")
	}

	return c.VoiceService().Stop(recipient, cid.String(), callID)
}

// Count returns the number of calls.
func (s service) Count(ctx context.Context, aID, cID string, callsSince int) (int, error) {
	return s.repo.Count(ctx, aID, cID, callsSince)
}

// Query returns the calls with the specified offset and limit.
func (s service) Query(ctx context.Context, aID, cID string, callsSince int, offset, limit int) ([]ExtCall, error) {
	items, err := s.repo.Query(ctx, aID, cID, callsSince, offset, limit)
	if err != nil {
		return nil, err
	}
	output := []ExtCall{}
	for _, i := range items {
		output = append(output, newExtCall(i))
	}
	return output, nil
}
