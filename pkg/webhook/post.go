package webhook

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
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
	TYPE_REQUEST      = "request"
	TYPE_RAW          = "raw"
	TYPE_VOICE_START  = "voice_start"
	TYPE_VOICE_BUSY   = "voice_busy"
	TYPE_VOICE_STOP   = "voice_stop"
	TYPE_VOICE_ACCEPT = "voice_accept"
	TYPE_VOICE_SETUP  = "voice_setup"
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
	Post(url, secret string, p WebhookPayload) error
}
type Webhook struct{}

func NewWebhook() *Webhook {
	return &Webhook{}
}

func (w Webhook) Post(url, secret string, p WebhookPayload) error {
	var postBody []byte
	var err error

	//Encode the data
	postBody, err = json.Marshal(p)
	if err != nil {
		return fmt.Errorf("error marshalling request: %v", err)
	}

	w.sendRequest(url, secret, postBody)
	return nil
}

// Function to compute HMAC hex digest
func (w Webhook) computeHMAC256(message, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}

func (w Webhook) sendRequest(callbackURL, secret string, responseBody []byte) error {
	// Create a new HTTP request
	req, err := http.NewRequest("POST", callbackURL, bytes.NewBuffer(responseBody))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	// Set the content header
	req.Header.Set("Content-Type", "application/json")

	// Set the HMAC hex digest signature header
	if len(secret) > 0 {
		signature := w.computeHMAC256(string(responseBody), secret)
		req.Header.Set("X-Hub-Signature-256", fmt.Sprintf("sha256=%s", signature))
	}

	// Send the request using an HTTP client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error when calling callback webhook: %v", err)
	}
	defer resp.Body.Close()

	return nil
}
