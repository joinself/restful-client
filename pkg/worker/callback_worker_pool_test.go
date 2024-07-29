package worker

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/joinself/restful-client/pkg/log"
	"github.com/joinself/restful-client/pkg/webhook"
	"github.com/maragudk/goqite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockQueueManager is a mock implementation of the QueueManager interface
type MockQueueManager struct {
	mock.Mock
}

func (m *MockQueueManager) Send(ctx context.Context, msg goqite.Message) error {
	args := m.Called(ctx, msg)
	return args.Error(0)
}

func (m *MockQueueManager) Receive(ctx context.Context) (*goqite.Message, error) {
	args := m.Called(ctx)
	msg := args.Get(0)
	if msg == nil {
		return nil, errors.New("hola")
	}
	return msg.(*goqite.Message), nil
}

func (m *MockQueueManager) Delete(ctx context.Context, id goqite.ID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockQueueManager) Extend(ctx context.Context, id goqite.ID, t time.Duration) error {
	args := m.Called(ctx, id, t)
	return args.Error(0)
}

// MockCallbackSender
type MockCallbackSender struct {
	mock.Mock
	Error error
}

func (m *MockCallbackSender) SendCallback(appID string, payload webhook.WebhookPayload) error {
	return m.Error
}

func TestCallbackWorkerPool_StartStop(t *testing.T) {
	payload := []byte(`{"app_id":"appID","webhook":{"typ":"typ","uri":"uri","data":"data","payload":{}}}`)
	mockQueue := new(MockQueueManager)
	mockQueue.On("Receive", mock.Anything).Return(&goqite.Message{
		ID:   "msg1",
		Body: payload,
	}, nil).Once()
	mockQueue.On("Receive", mock.Anything).Return(nil, nil)
	mockQueue.On("Delete", context.Background(), goqite.ID("msg1")).Return(nil)
	mockLogger, _ := log.NewForTest()
	mockCallbackSender := new(MockCallbackSender)
	mockCallbackSender.On("SendCallback", "lol", payload).Once()

	pool := NewCallbackWorkerPool(mockQueue, mockLogger, mockCallbackSender, 3)
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

func TestCallbackWorkerPool_DeleteOnError(t *testing.T) {
	mockQueue := new(MockQueueManager)
	mockQueue.On("Receive", mock.Anything).Return(&goqite.Message{
		ID:   "msg1",
		Body: []byte("{]"),
	}, nil).Once()
	mockQueue.On("Receive", mock.Anything).Return(nil, nil)
	mockQueue.On("Delete", context.Background(), goqite.ID("msg1")).Return(nil)
	mockLogger, _ := log.NewForTest()
	mockCallbackSender := new(MockCallbackSender)

	pool := NewCallbackWorkerPool(mockQueue, mockLogger, mockCallbackSender, 3)
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

func TestCallbackWorkerPool_ExtendOnSendingError(t *testing.T) {
	payload := []byte(`{"app_id":"appID","webhook":{"typ":"typ","uri":"uri","data":"data","payload":{}}}`)
	mockQueue := new(MockQueueManager)
	mockQueue.On("Receive", mock.Anything).Return(&goqite.Message{
		ID:   "msg1",
		Body: payload,
	}, nil).Once()
	mockQueue.On("Receive", mock.Anything).Return(nil, nil)
	mockQueue.On("Extend", context.Background(), goqite.ID("msg1"), mock.Anything).Return(nil)
	mockLogger, _ := log.NewForTest()
	mockCallbackSender := new(MockCallbackSender)
	mockCallbackSender.Error = errors.New("error calling back")
	mockCallbackSender.On("SendCallback", mock.Anything, mock.Anything).Return(errors.New("fail")).Once()

	pool := NewCallbackWorkerPool(mockQueue, mockLogger, mockCallbackSender, 3)
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
	mockQueue := new(MockQueueManager)
	mockQueue.On("Receive", mock.Anything).Return(nil, nil)
	mockLogger, _ := log.NewForTest()
	mockCallbackSender := new(MockCallbackSender)

	pool := NewCallbackWorkerPool(mockQueue, mockLogger, mockCallbackSender, 3)
	pool.Start()

	// Test data
	qm := CallbackTask{
		AppID: "test-app-id",
		// Callback:       "http://example.com/callback",
		// CallbackSecret: "secret",
		WebhookPayload: webhook.WebhookPayload{},
	}
	body, _ := json.Marshal(qm)

	mockQueue.On("Send", context.Background(), goqite.Message{Body: body}).Return(nil)

	err := pool.Send(qm)
	assert.NoError(t, err)

	mockQueue.AssertExpectations(t)
	pool.Stop()
}
