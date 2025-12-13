package shared

import (
	"context"
	"sync"
	"time"
)

type Observer[E any] struct {
	lock      sync.RWMutex
	topics    map[int64]map[string]bool
	observers map[string]chan *E
}

func NewObserver[E any]() *Observer[E] {
	return &Observer[E]{
		topics:    make(map[int64]map[string]bool),
		observers: make(map[string]chan *E),
	}
}

func (p *Observer[E]) Observe(ctx context.Context, observerID string, topicIDs ...int64) (<-chan *E, func()) {
	p.lock.Lock()
	defer p.lock.Unlock()

	for _, topicID := range topicIDs {
		topic, ok := p.topics[topicID]
		if !ok {
			topic = make(map[string]bool)
			p.topics[topicID] = topic
		}
		topic[observerID] = true
	}

	p.observers[observerID] = make(chan *E, 1)

	Logger.InfoContext(ctx, "observing", "observer_id", observerID, "topic_ids", topicIDs)
	return p.observers[observerID], p.cancel(ctx, observerID, topicIDs...)
}

func (p *Observer[E]) cancel(ctx context.Context, observerID string, topicIDs ...int64) func() {
	return func() {
		p.lock.Lock()
		defer p.lock.Unlock()

		for _, topicID := range topicIDs {
			topic, ok := p.topics[topicID]
			if ok {
				delete(topic, observerID)
			}
		}

		observer, ok := p.observers[observerID]
		if ok {
			delete(p.observers, observerID)
			close(observer)
		}

		Logger.InfoContext(ctx, "canceled observer", "observer_id", observerID, "topic_ids", topicIDs)
	}
}

func (p *Observer[E]) Notify(ctx context.Context, topicID int64, event *E) {
	p.lock.RLock()
	defer p.lock.RUnlock()

	for observerID := range p.topics[topicID] {
		go func() {
			select {
			case p.observers[observerID] <- event:
				Logger.InfoContext(ctx, "notified event", "event", event, "topic_id", topicID, "observer_id", observerID)
			case <-time.After(100 * time.Millisecond):
				Logger.ErrorContext(ctx, "failed to notify event", "event", event, "topic_id", topicID, "observer_id", observerID)
			}
		}()
	}
}
