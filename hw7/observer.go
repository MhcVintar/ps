package main

import (
	"hw7/protobufStorage"
	"sync"
	"time"
)

type Observer struct {
	lock      sync.RWMutex
	observers map[int]chan *protobufStorage.TodoEvent
	nextID    int
}

func NewObserver() *Observer {
	return &Observer{
		observers: make(map[int]chan *protobufStorage.TodoEvent),
	}
}

func (p *Observer) Observe() (<-chan *protobufStorage.TodoEvent, func()) {
	p.lock.Lock()
	defer p.lock.Unlock()

	observerID := p.nextID
	p.nextID++
	p.observers[observerID] = make(chan *protobufStorage.TodoEvent, 5)

	return p.observers[observerID], p.cancel(observerID)
}

func (p *Observer) cancel(observerID int) func() {
	return func() {
		p.lock.Lock()
		defer p.lock.Unlock()

		observer, ok := p.observers[observerID]
		if ok {
			delete(p.observers, observerID)
			close(observer)
		}
	}
}

func (p *Observer) Notify(event *protobufStorage.TodoEvent) {
	p.lock.RLock()
	defer p.lock.RUnlock()

	for _, observer := range p.observers {
		select {
		case observer <- event:
		case <-time.After(100 * time.Millisecond):
		}
	}
}
