package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/joinself/restful-client/pkg/log"
	"github.com/joinself/restful-client/pkg/webhook"
	"github.com/maragudk/goqite"
)

const (
	extendTimeout = 5 * time.Minute
)

// Task represents a unit of work to be processed.
type QueuedMessage struct {
	AppID          string                 `json:"app_id"`
	Callback       string                 `json:"callback"`
	CallbackSecret string                 `json:"callback_secret"`
	WebhookPayload webhook.WebhookPayload `json:"webhook"`
}

// Worker represents a single worker
type CallbackWorker struct {
	id     int
	queue  *goqite.Queue
	logger log.Logger
	quit   chan bool
}

// NewCallbackWorker creates a new worker
func NewCallbackWorker(id int, queue *goqite.Queue, logger log.Logger) CallbackWorker {
	return CallbackWorker{
		id:     id,
		queue:  queue,
		logger: logger,
		quit:   make(chan bool),
	}
}

// Start begins the worker's task processing loop
func (w CallbackWorker) Start(wg *sync.WaitGroup) {
	go func() {
		defer wg.Done()
		for {
			select {
			case <-w.quit:
				return
			default:
				if item, err := w.queue.Receive(context.Background()); err == nil && item != nil {
					fmt.Printf("Worker %d processing task: %s\n", w.id, item.ID)
					w.processTask(item)
				} else {
					time.Sleep(100 * time.Millisecond) // Avoid busy waiting
				}
			}
		}
	}()
}

// Stop signals the worker to stop
func (w CallbackWorker) Stop() {
	go func() {
		w.quit <- true
	}()
}

func (w *CallbackWorker) processTask(m *goqite.Message) error {
	var qm QueuedMessage
	err := json.Unmarshal(m.Body, &qm)
	if err != nil {
		w.logger.Error("error unmarshalling queued message, deleting message")
		return w.queue.Delete(context.Background(), m.ID)
	}

	err = webhook.NewWebhook().Post(qm.Callback, qm.CallbackSecret, qm.WebhookPayload)
	if err != nil {
		w.logger.Infof("extending queueed message %s : %s", m.ID, err.Error())
		if err := w.queue.Extend(context.Background(), m.ID, extendTimeout); err != nil {
			w.logger.Error("error extending message timeout")
		}
		return err
	}

	w.logger.Infof("deleting queueed message %s", m.ID)
	return w.queue.Delete(context.Background(), m.ID)
}

// CallbackWorkerPool manages the task queue and worker pool
type CallbackWorkerPool struct {
	queue   *goqite.Queue
	logger  log.Logger
	workers []CallbackWorker
	wg      sync.WaitGroup
}

// NewCallbackWorkerPool creates a new worker pool
func NewCallbackWorkerPool(queue *goqite.Queue, logger log.Logger, numWorkers int) *CallbackWorkerPool {
	return &CallbackWorkerPool{
		queue:   queue,
		logger:  logger,
		workers: make([]CallbackWorker, numWorkers),
	}
}

// Start initializes and starts the workers
func (wp *CallbackWorkerPool) Start() {
	for i := 0; i < len(wp.workers); i++ {
		worker := NewCallbackWorker(i+1, wp.queue, wp.logger)
		wp.workers[i] = worker
		wp.wg.Add(1)
		worker.Start(&wp.wg)
	}
}

// Send adds a task to the task queue
func (wp *CallbackWorkerPool) Send(qm QueuedMessage) error {
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
