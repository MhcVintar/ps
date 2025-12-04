package shared

import (
	"context"
	"sync"
	"time"
)

type Observer[E any] struct {
	lock          sync.RWMutex
	topics        map[int64]map[string]bool
	subscriptions map[string]chan *E
}

func NewObserver[E any]() *Observer[E] {
	return &Observer[E]{
		topics:        make(map[int64]map[string]bool),
		subscriptions: make(map[string]chan *E),
	}
}

func (p *Observer[E]) Subscribe(ctx context.Context, subscriberID string, topicIDs ...int64) <-chan *E {
	p.lock.Lock()
	defer p.lock.Unlock()

	for _, topicID := range topicIDs {
		topic, ok := p.topics[topicID]
		if !ok {
			topic = make(map[string]bool)
		}
		topic[subscriberID] = true
	}

	p.subscriptions[subscriberID] = make(chan *E, 1)

	Logger.InfoContext(ctx, "subscribed", "subscriber_id", subscriberID, "topic_ids", topicIDs)
	return p.subscriptions[subscriberID]
}

func (p *Observer[E]) Unsubscribe(ctx context.Context, subscriberID string, topicIDs ...int64) {
	p.lock.Lock()
	defer p.lock.Unlock()

	for _, topicID := range topicIDs {
		topic, ok := p.topics[topicID]
		if ok {
			delete(topic, subscriberID)
		}
	}

	subscription, ok := p.subscriptions[subscriberID]
	if ok {
		delete(p.subscriptions, subscriberID)
		close(subscription)
	}

	Logger.InfoContext(ctx, "unsubscribed", "subscriber_id", subscriberID, "topic_ids", topicIDs)
}

func (p *Observer[E]) Publish(ctx context.Context, topicID int64, event *E) {
	// TODO Make sure to set the event at before calling the publish
	p.lock.RLock()
	defer p.lock.RUnlock()

	for subscriberID := range p.topics[topicID] {
		go func(subscription chan *E) {
			select {
			case subscription <- event:
				Logger.InfoContext(ctx, "event published", "event", event, "topic_id", topicID, "subscriber_id", subscriberID)
			case <-time.After(100 * time.Millisecond):
				Logger.ErrorContext(ctx, "failed to publish event", "event", event, "topic_id", topicID, "subscriber_id", subscriberID)
			}
		}(p.subscriptions[subscriberID])
	}
}
