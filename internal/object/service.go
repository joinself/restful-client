package object

import (
	"context"
	"encoding/base64"
	"errors"
	"strings"

	"github.com/joinself/restful-client/pkg/log"
	"github.com/joinself/restful-client/pkg/support"
)

// Service encapsulates usecase logic for signatures.
type Service interface {
	BuildObject(ctx context.Context, appID string, o Object) (*ExtObject, error)
}

// CreateSignatureRequest represents an signature creation request.
type service struct {
	runner support.SelfClientGetter
	logger log.Logger
}

// NewService creates a new signature service.
func NewService(runner support.SelfClientGetter, logger log.Logger) Service {
	return service{runner, logger}
}

func (s service) BuildObject(ctx context.Context, appID string, o Object) (*ExtObject, error) {
	// Check there is a runner for this app
	client, ok := s.runner.Get(appID)
	if !ok {
		return nil, errors.New("app not found")
	}

	input := o.DataURI
	b64data := input[strings.IndexByte(input, ',')+1:]
	content, err := base64.RawStdEncoding.DecodeString(b64data)
	mime := input[strings.IndexByte(input, ':')+1 : strings.Index(input, ";")]

	obj := client.ChatService().NewObject()
	err = obj.BuildFromData(content, "object", mime)
	if err != nil {
		s.logger.Error(err.Error())
		return nil, errors.New("problem building object from data")
	}

	return &ExtObject{
		Link:    obj.Link,
		Name:    obj.Name,
		Mime:    obj.Mime,
		Expires: obj.Expires,
		Key:     obj.Key,
	}, nil

}
