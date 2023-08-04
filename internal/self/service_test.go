package self

import (
	"encoding/json"
	"testing"

	"github.com/joinself/restful-client/internal/entity"
	"github.com/joinself/restful-client/pkg/log"
	"github.com/joinself/restful-client/pkg/mock"
	"github.com/joinself/restful-client/pkg/webhook"
	"github.com/joinself/self-go-sdk/messaging"
	"github.com/stretchr/testify/assert"
)

type config struct {
	mRepo *mock.MessageRepositoryMock
	cRepo *mock.ConnectionRepositoryMock
	fRepo *mock.FactRepositoryMock
	wMock *mock.PosterMock
	sMock *mock.SelfMock
}

func buildService(c *config) Service {
	logger, _ := log.NewForTest()
	if c.mRepo == nil {
		c.mRepo = &mock.MessageRepositoryMock{}
	}
	if c.cRepo == nil {
		c.cRepo = &mock.ConnectionRepositoryMock{}
	}
	if c.fRepo == nil {
		c.fRepo = &mock.FactRepositoryMock{}
	}
	if c.sMock == nil {
		c.sMock = &mock.SelfMock{}
	}
	if c.wMock == nil {
		c.wMock = &mock.PosterMock{}
	}

	return NewService(c.sMock, c.cRepo, c.fRepo, c.mRepo, logger, c.wMock)

}

func TestProcessFactsQueryResp(t *testing.T) {
	c := config{}
	s := buildService(&c)

	var payload map[string]interface{}
	var ExportProcessQueryResp = (Service).processFactsQueryResp
	ExportProcessQueryResp(s, payload)

	last := c.wMock.History[len(c.wMock.History)-1]
	assert.Equal(t, webhook.TYPE_RAW, last.Type)
	assert.Equal(t, "", last.URI)
	assert.Equal(t, payload, last.Data)
}

func TestProcessChatMessage(t *testing.T) {
	c := config{}
	s := buildService(&c)

	payload := map[string]interface{}{
		"iss": "ISS",
		"msg": "MSG",
		"jti": "JTI",
		"aud": "AUD",
	}
	var ExportProcessChatMessage = (Service).processChatMessage
	ExportProcessChatMessage(s, payload)

	last := c.wMock.History[len(c.wMock.History)-1]
	assert.Equal(t, webhook.TYPE_MESSAGE, last.Type)
	assert.Equal(t, "/apps/test/connections/ISS/messages/JTI", last.URI)
	data := last.Data.(entity.Message)
	assert.Equal(t, 0, data.ConnectionID)
	assert.Equal(t, "", data.CID)
	assert.Equal(t, "JTI", data.JTI)
	assert.Equal(t, "", data.RID)
	assert.Equal(t, "MSG", data.Body)

	// Check a new connection has been created
	assert.Equal(t, 1, len(c.cRepo.Items))

	lastConnection := c.cRepo.Items[len(c.cRepo.Items)-1]
	assert.Equal(t, "ISS", lastConnection.SelfID)
	assert.Equal(t, "test", lastConnection.AppID)
	assert.Equal(t, "-", lastConnection.Name)

	// Check a message has been created
	assert.Equal(t, 1, len(c.mRepo.Items))

	lastMsg := c.mRepo.Items[len(c.mRepo.Items)-1]
	assert.Equal(t, "ISS", lastMsg.ISS)
	assert.Equal(t, 0, lastMsg.ConnectionID)
	assert.Equal(t, "", lastMsg.CID)
	assert.Equal(t, "JTI", lastMsg.JTI)
	assert.Equal(t, "", lastMsg.RID)
	assert.Equal(t, "MSG", lastMsg.Body)
}

func TestProcessConnectionResp(t *testing.T) {
	c := config{}
	s := buildService(&c)

	payload := map[string]interface{}{
		"iss": "ISS",
		"msg": "MSG",
		"jti": "JTI",
		"aud": "AUD",
	}
	var ExportProcessConnectionResp = (Service).processConnectionResp
	ExportProcessConnectionResp(s, payload)

	last := c.wMock.History[len(c.wMock.History)-1]
	assert.Equal(t, webhook.TYPE_CONNECTION, last.Type)
	assert.Equal(t, "/apps/test/connections/ISS", last.URI)
	data := last.Data.(entity.Connection)
	assert.Equal(t, 0, data.ID)
	assert.Equal(t, "ISS", data.SelfID)
	assert.Equal(t, "test", data.AppID)
	assert.Equal(t, "-", data.Name)

	// Check a new connection has been created
	assert.Equal(t, 1, len(c.cRepo.Items))

	lastConnection := c.cRepo.Items[len(c.cRepo.Items)-1]
	assert.Equal(t, "ISS", lastConnection.SelfID)
	assert.Equal(t, "test", lastConnection.AppID)
	assert.Equal(t, "-", lastConnection.Name)
}

func TestProcessConnectionRespWithName(t *testing.T) {
	c := config{}
	s := buildService(&c)

	payload := map[string]interface{}{
		"iss": "ISS",
		"msg": "MSG",
		"jti": "JTI",
		"aud": "AUD",
		"data": map[string]string{
			"name": "NAME",
		},
	}
	var ExportProcessConnectionResp = (Service).processConnectionResp
	ExportProcessConnectionResp(s, payload)

	last := c.wMock.History[len(c.wMock.History)-1]
	assert.Equal(t, webhook.TYPE_CONNECTION, last.Type)
	assert.Equal(t, "/apps/test/connections/ISS", last.URI)
	data := last.Data.(entity.Connection)
	assert.Equal(t, 0, data.ID)
	assert.Equal(t, "ISS", data.SelfID)
	assert.Equal(t, "test", data.AppID)
	assert.Equal(t, "NAME", data.Name)

	// Check a new connection has been created
	assert.Equal(t, 1, len(c.cRepo.Items))

	lastConnection := c.cRepo.Items[len(c.cRepo.Items)-1]
	assert.Equal(t, "ISS", lastConnection.SelfID)
	assert.Equal(t, "test", lastConnection.AppID)
	assert.Equal(t, "NAME", lastConnection.Name)
}

func TestProcessIncomingMessage(t *testing.T) {
	c := config{}
	s := buildService(&c)

	tests := map[string]string{
		"chat.message":                webhook.TYPE_MESSAGE,
		"identities.connections.resp": webhook.TYPE_CONNECTION,
		"identities.facts.query.resp": webhook.TYPE_RAW,
	}

	for typ, result := range tests {
		payload, err := json.Marshal(map[string]interface{}{
			"typ": typ,
			"iss": "ISS",
			"msg": "MSG",
			"jti": "JTI",
			"aud": "AUD",
			"data": map[string]string{
				"name": "NAME",
			},
		})
		assert.NoError(t, err)

		m := messaging.Message{Payload: payload}
		var ExportProcessIncomingMessage = (Service).processIncomingMessage
		ExportProcessIncomingMessage(s, &m)

		last := c.wMock.History[len(c.wMock.History)-1]
		assert.Equal(t, result, last.Type)
	}

}

func TestEnsureSelfClientIsStarted(t *testing.T) {
	c := config{}
	s := buildService(&c)

	assert.False(t, c.sMock.Started)
	s.Run()
	assert.True(t, c.sMock.Started)
}
