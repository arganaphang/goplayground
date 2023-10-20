package pubsub

import "sync"

type pubsub struct {
	mu          sync.RWMutex
	subscribers map[string][]chan interface{}
	closed      bool
}

func New() *pubsub {
	ps := &pubsub{}
	ps.subscribers = make(map[string][]chan interface{})
	return ps
}

func (ps *pubsub) Subscribe(topic string) <-chan interface{} {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ch := make(chan interface{}, 1)
	ps.subscribers[topic] = append(ps.subscribers[topic], ch)
	return ch
}

func (ps *pubsub) Publish(topic string, data interface{}) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	if ps.closed {
		return
	}

	for _, ch := range ps.subscribers[topic] {
		go func(ch chan interface{}) {
			ch <- data
		}(ch)
	}
}

func (ps *pubsub) Close() {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	if !ps.closed {
		ps.closed = true
		for _, subscribers := range ps.subscribers {
			for _, ch := range subscribers {
				close(ch)
			}
		}
	}
}
