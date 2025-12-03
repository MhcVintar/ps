package shared

import (
	"context"
	"razpravljalnica/internal/api"
	"sync"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type PubSub struct {
	lock          sync.RWMutex
	topics        map[int64]map[string]bool
	subscriptions map[string]chan *api.MessageEvent
}

func NewPubSub() *PubSub {
	return &PubSub{
		topics:        make(map[int64]map[string]bool),
		subscriptions: make(map[string]chan *api.MessageEvent),
	}
}

func (p *PubSub) Subscribe(ctx context.Context, subscriberID string, topicIDs ...int64) <-chan *api.MessageEvent {
	p.lock.Lock()
	defer p.lock.Unlock()

	for _, topicID := range topicIDs {
		topic, ok := p.topics[topicID]
		if !ok {
			topic = make(map[string]bool)
		}
		topic[subscriberID] = true
	}

	p.subscriptions[subscriberID] = make(chan *api.MessageEvent, 1)

	Logger.InfoContext(ctx, "subscribed", "subscriber_id", subscriberID, "topic_ids", topicIDs)
	return p.subscriptions[subscriberID]
}

func (p *PubSub) Unsubscribe(ctx context.Context, subscriberID string, topicIDs ...int64) {
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

func (p *PubSub) Publish(ctx context.Context, topicID int64, event *api.MessageEvent) {
	event.EventAt = timestamppb.Now()

	p.lock.RLock()
	defer p.lock.RUnlock()

	for subscriberID := range p.topics[topicID] {
		go func(subscription chan *api.MessageEvent) {
			select {
			case subscription <- event:
				Logger.InfoContext(ctx, "event published", "event", event, "topic_id", topicID, "subscriber_id", subscriberID)
			case <-time.After(100 * time.Millisecond):
				Logger.ErrorContext(ctx, "failed to publish event", "event", event, "topic_id", topicID, "subscriber_id", subscriberID)
			}
		}(p.subscriptions[subscriberID])
	}
}
