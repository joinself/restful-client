package worker

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/joinself/restful-client/pkg/log"
	"github.com/maragudk/goqite"
)

const (
	extendTimeout = 5 * time.Minute
)

// Worker represents a single worker
type CallbackWorker struct {
	id             int
	queue          *goqite.Queue
	logger         log.Logger
	callbackSender CallbackSender
	quit           chan bool
}

// NewCallbackWorker creates a new worker
func NewCallbackWorker(id int, queue *goqite.Queue, logger log.Logger, callbackSender CallbackSender) CallbackWorker {
	return CallbackWorker{
		id:             id,
		queue:          queue,
		logger:         logger,
		callbackSender: callbackSender,
		quit:           make(chan bool),
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
					w.logger.Infof("Worker %d processing task: %s\n", w.id, item.ID)
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
	var t CallbackTask
	err := json.Unmarshal(m.Body, &t)
	if err != nil {
		w.logger.Error("error unmarshalling task, deleting message")
		return w.queue.Delete(context.Background(), m.ID)
	}

	if err = t.Send(w.callbackSender); err != nil {
		w.logger.Infof("extending task %s : %s", m.ID, err.Error())
		if err := w.queue.Extend(context.Background(), m.ID, extendTimeout); err != nil {
			w.logger.Error("error extending task timeout")
		}
		return err
	}
	w.logger.Infof("deleting task %s", m.ID)
	return w.queue.Delete(context.Background(), m.ID)
}
