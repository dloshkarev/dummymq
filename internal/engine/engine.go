package engine

import (
	. "dummymq/internal/model"
	"fmt"
	"sync"
	"time"
)

type Engine struct {
	queues      map[string]chan *Message
	subscribers map[string]map[string]chan *Message
	stopFlags   *sync.Map
	mu          *sync.Mutex
}

func Run(queueNames []string, messageLimit int) *Engine {
	subscribers := make(map[string]map[string]chan *Message)
	var stopFlags sync.Map

	queues := make(map[string]chan *Message)
	for _, q := range queueNames {
		queues[q] = make(chan *Message, messageLimit)
		stopFlags.Store(q, false)
	}

	var mutex sync.Mutex
	return &Engine{queues: queues, subscribers: subscribers, stopFlags: &stopFlags, mu: &mutex}
}

func (e Engine) AddMessage(queue string, msg *Message) error {
	go func() {
		e.queues[queue] <- msg
	}()
	return nil
}

func (e Engine) Subscribe(queue string, endpoint string, messageLimit int) (chan *Message, error) {
	defer e.mu.Unlock()
	e.mu.Lock()

	msgs, ok := e.queues[queue]
	if ok {
		var outCh chan *Message
		outCh, exists := e.subscribers[queue][endpoint]

		if !exists {
			e.stopFlags.Store(queue, true)
			outCh = make(chan *Message, messageLimit)

			if _, exists := e.subscribers[queue]; !exists {
				e.subscribers[queue] = make(map[string]chan *Message)
			}
			e.subscribers[queue][endpoint] = outCh

			go func() {
				for {
					select {
					case msg := <-msgs:
						checkStopReading(&e, queue)
						for _, ch := range e.subscribers[queue] {
							ch <- msg
						}
					default:
						checkStopReading(&e, queue)
						// avoid cpu burn
						time.Sleep(100 * time.Millisecond)
					}
				}
			}()
		}
		return outCh, nil
	} else {
		return nil, fmt.Errorf("queue %v doesn't exists", queue)
	}
}

func checkStopReading(e *Engine, queue string) {
	if stop, ok := e.stopFlags.Load(queue); ok && stop.(bool) {
		e.stopFlags.Store(queue, false)
		return
	}
}
