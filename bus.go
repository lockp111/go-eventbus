package eventbus

import (
	"reflect"
	"sync"
)

// Bus struct
type Bus[T any] struct {
	mux    sync.Mutex
	events map[string][]*event[T]
}

// New - return a new Bus object
func New[T any]() *Bus[T] {
	return &Bus[T]{
		events: make(map[string][]*event[T]),
	}
}

// On - register topic event and return error
func (b *Bus[T]) On(topic string, e ...Event[T]) *Bus[T] {
	b.addEvents(topic, false, e)
	return b
}

// Once - register once event and return error
func (b *Bus[T]) Once(topic string, e ...Event[T]) *Bus[T] {
	b.addEvents(topic, true, e)
	return b
}

// Off - remove topic event
func (b *Bus[T]) Off(topic string, e ...Event[T]) *Bus[T] {
	b.removeEvents(topic, e)
	return b
}

// Clean - clear all events
func (b *Bus[T]) Clean() *Bus[T] {
	b.mux.Lock()
	defer b.mux.Unlock()
	b.events = make(map[string][]*event[T])
	return b
}

// Trigger - dispatch event
func (b *Bus[T]) Trigger(topic string, msg ...T) *Bus[T] {
	b.dispatch(topic, msg)
	return b
}

func (b *Bus[T]) addEvents(topic string, isUnique bool, es []Event[T]) {
	if len(es) == 0 {
		return
	}

	b.mux.Lock()
	defer b.mux.Unlock()

	for _, e := range es {
		b.events[topic] = append(b.events[topic], newEvent(e, topic, isUnique))
	}
}

func (b *Bus[T]) removeEvents(topic string, es []Event[T]) {
	b.mux.Lock()
	defer b.mux.Unlock()

	if len(es) == 0 {
		delete(b.events, topic)
		return
	}

	events := b.events[topic]
	if len(events) == 0 {
		return
	}

	for _, e := range es {
		tag := reflect.ValueOf(e)
		for i := 0; i < len(events); i++ {
			if events[i].tag == tag {
				events = append(events[:i], events[i+1:]...)
				i--
			}
		}
	}

	if len(events) == 0 {
		delete(b.events, topic)
		return
	}

	b.events[topic] = events
}

func (b *Bus[T]) getEvents(topic string) []*event[T] {
	b.mux.Lock()
	defer b.mux.Unlock()

	events := make([]*event[T], 0, len(b.events[topic])+len(b.events[ALL]))
	for _, e := range b.events[topic] {
		if e.isUnique {
			if e.hasCalled {
				continue
			}
			e.hasCalled = true
		}
		events = append(events, e)
	}

	if topic != ALL {
		for _, e := range b.events[ALL] {
			if e.isUnique {
				if e.hasCalled {
					continue
				}
				e.hasCalled = true
			}
			events = append(events, e)
		}
	}
	return events
}

func (b *Bus[T]) dispatch(topic string, data []T) {
	var (
		events  = b.getEvents(topic)
		removes = make(map[string][]Event[T])
	)

	for _, e := range events {
		e.Dispatch(data...)
		if e.isUnique && e.hasCalled {
			removes[e.topic] = append(removes[e.topic], e.Event)
		}
	}

	for k, v := range removes {
		b.removeEvents(k, v)
	}
}
