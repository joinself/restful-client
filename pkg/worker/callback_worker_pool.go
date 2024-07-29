package worker

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/joinself/restful-client/pkg/log"
	"github.com/joinself/restful-client/pkg/webhook"
	"github.com/maragudk/goqite"
)

type CallbackSender interface {
	SendCallback(appID string, payload webhook.WebhookPayload) error
}

// CallbackWorkerPool manages the task queue and worker pool
type CallbackWorkerPool struct {
	queue          QueueManager
	logger         log.Logger
	workers        []CallbackWorker
	wg             sync.WaitGroup
	callbackSender CallbackSender
}

// NewCallbackWorkerPool creates a new worker pool
func NewCallbackWorkerPool(queue QueueManager, logger log.Logger, callbackSender CallbackSender, numWorkers int) *CallbackWorkerPool {
	return &CallbackWorkerPool{
		queue:          queue,
		logger:         logger,
		callbackSender: callbackSender,
		workers:        make([]CallbackWorker, numWorkers),
	}
}

// Start initializes and starts the workers
func (wp *CallbackWorkerPool) Start() {
	for i := 0; i < len(wp.workers); i++ {
		worker := NewCallbackWorker(i+1, wp.queue, wp.logger, wp.callbackSender)
		wp.workers[i] = worker
		wp.wg.Add(1)
		worker.Start(&wp.wg)
	}
}

// Send adds a task to the task queue
func (wp *CallbackWorkerPool) Send(qm CallbackTask) error {
	body, err := json.Marshal(qm)

	if err != nil {
		return err
	}

	return wp.queue.Send(context.Background(), goqite.Message{
		Body: body,
	})
}

// Stop signals all workers to stop
func (wp *CallbackWorkerPool) Stop() {
	for _, worker := range wp.workers {
		worker.Stop()
	}
	wp.wg.Wait()
}
