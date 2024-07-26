package worker

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/joinself/restful-client/pkg/webhook"
	"github.com/maragudk/goqite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockQueue is a mock implementation of the goqite.Queue interface
type MockQueue struct {
	mock.Mock
}

func (m *MockQueue) Send(ctx context.Context, msg goqite.Message) error {
	args := m.Called(ctx, msg)
	return args.Error(0)
}

func (m *MockQueue) Receive(ctx context.Context) (*goqite.Message, error) {
	args := m.Called(ctx)
	return args.Get(0).(*goqite.Message), args.Error(1)
}

func (m *MockQueue) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockQueue) Extend(ctx context.Context, id string, t time.Duration) error {
	args := m.Called(ctx, id, t)
	return args.Error(0)
}

// MockLogger is a mock implementation of the log.Logger interface
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Error(args ...interface{}) {
	m.Called(args...)
}

func (m *MockLogger) Infof(format string, args ...interface{}) {
	m.Called(append([]interface{}{format}, args...)...)
}

// MockWebhook is a mock implementation of the webhook.Webhook interface
type MockWebhook struct {
	mock.Mock
}

func (m *MockWebhook) Post(url, secret string, payload webhook.WebhookPayload) error {
	args := m.Called(url, secret, payload)
	return args.Error(0)
}

func TestCallbackWorker_ProcessTask(t *testing.T) {
	mockQueue := new(MockQueue)
	mockLogger := new(MockLogger)
	mockWebhook := new(MockWebhook)

	webhook.NewWebhook = func() webhook.Webhook {
		return mockWebhook
	}

	worker := NewCallbackWorker(1, mockQueue, mockLogger)

	// Test data
	qm := QueuedMessage{
		AppID:          "test-app-id",
		Callback:       "http://example.com/callback",
		CallbackSecret: "secret",
		WebhookPayload: webhook.WebhookPayload{},
	}
	body, _ := json.Marshal(qm)
	message := &goqite.Message{
		ID:   "message-id",
		Body: body,
	}

	// Test successful processing
	mockWebhook.On("Post", "http://example.com/callback", "secret", qm.WebhookPayload).Return(nil)
	mockQueue.On("Delete", context.Background(), "message-id").Return(nil)
	mockLogger.On("Infof", mock.Anything, mock.Anything).Return()

	err := worker.processTask(message)
	assert.NoError(t, err)

	mockWebhook.AssertExpectations(t)
	mockQueue.AssertExpectations(t)
	mockLogger.AssertExpectations(t)

	// Test webhook post failure
	mockWebhook.On("Post", "http://example.com/callback", "secret", qm.WebhookPayload).Return(errors.New("webhook error"))
	mockQueue.On("Extend", context.Background(), "message-id", extendTimeout).Return(nil)
	mockLogger.On("Infof", mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything).Return()

	err = worker.processTask(message)
	assert.Error(t, err)

	mockWebhook.AssertExpectations(t)
	mockQueue.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestCallbackWorkerPool_StartStop(t *testing.T) {
	mockQueue := new(MockQueue)
	mockLogger := new(MockLogger)

	pool := NewCallbackWorkerPool(mockQueue, mockLogger, 3)
	pool.Start()

	// Ensure all workers are started
	assert.Equal(t, 3, len(pool.workers))

	time.Sleep(1 * time.Second) // Allow some time for workers to start

	// Stop all workers
	pool.Stop()

	// Ensure all workers are stopped
	for _, worker := range pool.workers {
		select {
		case <-worker.quit:
		default:
			t.Errorf("worker %d was not stopped", worker.id)
		}
	}
}

func TestCallbackWorkerPool_Send(t *testing.T) {
	mockQueue := new(MockQueue)
	mockLogger := new(MockLogger)

	pool := NewCallbackWorkerPool(mockQueue, mockLogger, 3)
	pool.Start()

	// Test data
	qm := QueuedMessage{
		AppID:          "test-app-id",
		Callback:       "http://example.com/callback",
		CallbackSecret: "secret",
		WebhookPayload: webhook.WebhookPayload{},
	}
	body, _ := json.Marshal(qm)

	mockQueue.On("Send", context.Background(), goqite.Message{Body: body}).Return(nil)

	err := pool.Send(qm)
	assert.NoError(t, err)

	mockQueue.AssertExpectations(t)
	pool.Stop()
}
