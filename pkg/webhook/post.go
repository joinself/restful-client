package webhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	// TYPE_MESSAGE webhook type used when a message is received
	TYPE_MESSAGE = "message"
	// TYPE_FACT_RESPONSE webhook type used when an untracked fact response is received
	TYPE_FACT_RESPONSE = "fact_response"
	// TYPE_CONNECTION webhook type used when a connection is received
	TYPE_CONNECTION = "connection"
	// TYPE_REQUEST webhook type used when a request response is received
	TYPE_REQUEST = "request"
	TYPE_RAW     = "raw"
)

// WebhookPayload represents a the payload that will be resent to the
// configured webhook URL if provided.
type WebhookPayload struct {
	// Type is the type of the message.
	Type string `json:"typ"`
	// URI is the URI you can fetch more information about the object on the data field.
	URI string `json:"uri"`
	// Data the object to be sent.
	Data interface{} `json:"data"`
	// Payload the response payload received.
	Payload map[string]interface{} `json:"payload,omitempty"`
}

type Poster interface {
	Post(p WebhookPayload) error
}
type Webhook struct {
	callbackURL string
}

func NewWebhook(url string) *Webhook {
	return &Webhook{
		callbackURL: url,
	}
}

func (w Webhook) Post(p WebhookPayload) error {
	return w.post(p)
}

func (w Webhook) post(m interface{}) error {
	var postBody []byte
	var err error

	//Encode the data
	switch pb := m.(type) {
	case []byte:
		postBody = pb
	default:
		postBody, err = json.Marshal(m)
		if err != nil {
			return fmt.Errorf("error marshalling request: %v", err)
		}
	}

	responseBody := bytes.NewBuffer(postBody)

	//Leverage Go's HTTP Post function to make request
	_, err = http.Post(w.callbackURL, "application/json", responseBody)
	if err != nil {
		return fmt.Errorf("error when calling callback webhook %v", err)
	}
	return nil
}
