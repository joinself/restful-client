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
	"github.com/stretchr/testify/require"
)

type config struct {
	mRepo  *mock.MessageRepositoryMock
	cRepo  *mock.ConnectionRepositoryMock
	fRepo  *mock.FactRepositoryMock
	wMock  *mock.PosterMock
	sMock  *mock.SelfMock
	rRepo  *mock.RequestRepositoryMock
	rsMock *RequestServiceMock
	cwMock *mock.CallbackWorkerPoolMock
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
	if c.rRepo == nil {
		c.rRepo = &mock.RequestRepositoryMock{}
	}
	if c.sMock == nil {
		c.sMock = &mock.SelfMock{}
	}
	if c.wMock == nil {
		c.wMock = &mock.PosterMock{}
	}
	if c.rsMock == nil {
		c.rsMock = &RequestServiceMock{}
	}
	if c.cwMock == nil {
		c.cwMock = &mock.CallbackWorkerPoolMock{}
	}

	return NewService(Config{
		SelfClient:         c.sMock,
		ConnectionRepo:     c.cRepo,
		FactRepo:           c.fRepo,
		MessageRepo:        c.mRepo,
		RequestRepo:        c.rRepo,
		Logger:             logger,
		Poster:             c.wMock,
		RequestService:     c.rsMock,
		CallbackWorkerPool: c.cwMock,
	})

}

func TestProcessFactsQueryResp(t *testing.T) {
	c := config{}
	s := buildService(&c)
	s.SetApp(entity.App{
		ID:       "id",
		Callback: "http://localhost",
	})

	body := []byte(`{"facts":[]}`)
	payload := map[string]interface{}{
		"iss": "ISS",
		"sub": "SUB",
		"cid": "CID",
	}
	var ExportProcessQueryResp = (Service).processFactsQueryResp
	err := ExportProcessQueryResp(s, body, payload)
	require.NoError(t, err)

	last := c.cwMock.History[len(c.cwMock.History)-1]
	assert.Equal(t, webhook.TYPE_FACT_RESPONSE, last.Type)
	assert.Equal(t, "", last.URI)
	resp := last.Data.(entity.Response)
	assert.Equal(t, 0, len(resp.Facts))
}

func TestProcessChatMessage(t *testing.T) {
	c := config{}
	s := buildService(&c)
	s.SetApp(entity.App{
		ID:       "id",
		Callback: "http://localhost",
	})

	payload := map[string]interface{}{
		"iss": "ISS",
		"msg": "MSG",
		"jti": "JTI",
		"aud": "AUD",
	}
	var ExportProcessChatMessage = (Service).processChatMessage
	ExportProcessChatMessage(s, payload)

	last := c.cwMock.History[len(c.cwMock.History)-1]
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
	s.SetApp(entity.App{
		ID:       "id",
		Callback: "http://localhost",
	})

	payload := map[string]interface{}{
		"iss": "ISS",
		"msg": "MSG",
		"jti": "JTI",
		"aud": "AUD",
	}
	var ExportProcessConnectionResp = (Service).processConnectionResp
	ExportProcessConnectionResp(s, payload)

	last := c.cwMock.History[len(c.cwMock.History)-1]
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
	s.SetApp(entity.App{
		ID:       "id",
		Callback: "http://localhost",
	})

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

	last := c.cwMock.History[len(c.cwMock.History)-1]
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
	s.SetApp(entity.App{
		ID:       "id",
		Callback: "http://localhost",
	})

	tests := map[string]string{
		"chat.message":                webhook.TYPE_MESSAGE,
		"identities.connections.resp": webhook.TYPE_CONNECTION,
		"identities.facts.query.resp": webhook.TYPE_FACT_RESPONSE,
	}

	for typ, result := range tests {
		payload, err := json.Marshal(map[string]interface{}{
			"typ":    typ,
			"iss":    "ISS",
			"cid":    "CID",
			"sub":    "SUB",
			"msg":    "MSG",
			"jti":    "JTI",
			"aud":    "AUD",
			"status": "accepted",
			"data": map[string]string{
				"name": "NAME",
			},
		})
		assert.NoError(t, err)

		m := messaging.Message{Payload: payload}
		var ExportProcessIncomingMessage = (Service).processIncomingMessage
		ExportProcessIncomingMessage(s, &m)

		last := c.cwMock.History[len(c.cwMock.History)-1]
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

func TestProcessChatMesageRead(t *testing.T) {
	c := config{}
	s := buildService(&c)

	tests := map[bool][]interface{}{
		true:  []interface{}{},
		false: []interface{}{"cid"},
	}

	for expectedError, cids := range tests {
		payload := map[string]interface{}{
			"typ":    "chat.message.read",
			"iss":    "ISS",
			"msg":    "MSG",
			"jti":    "JTI",
			"aud":    "AUD",
			"status": "accepted",
			"cids":   cids,
		}
		var ExportProcessReadMessage = (Service).processChatMessageRead
		err := ExportProcessReadMessage(s, payload)
		if expectedError {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}
	}

}

func TestProcessChatMesageReceived(t *testing.T) {
	c := config{}
	s := buildService(&c)

	tests := map[bool][]interface{}{
		true:  []interface{}{},
		false: []interface{}{"cid"},
	}

	for expectedError, cids := range tests {
		payload := map[string]interface{}{
			"typ":    "chat.message.received",
			"iss":    "ISS",
			"msg":    "MSG",
			"jti":    "JTI",
			"aud":    "AUD",
			"status": "accepted",
			"cids":   cids,
		}
		var ExportProcessDeliveredMessage = (Service).processChatMessageDelivered
		err := ExportProcessDeliveredMessage(s, payload)
		if expectedError {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}
	}

}
