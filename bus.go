package eventbus

import (
	"reflect"
	"sync/atomic"

	"github.com/lockp111/go-cmap"
)

// Bus struct
type Bus[T any] struct {
	events cmap.ConcurrentMap[string, []*event[T]]
}

// New - return a new Bus object
func New[T any]() *Bus[T] {
	return &Bus[T]{
		events: cmap.New[[]*event[T]](),
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
	b.events = cmap.New[[]*event[T]]()
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
	for _, e := range es {
		b.events.Upsert(topic, func(oldValue []*event[T], exist bool) []*event[T] {
			return append(oldValue, newEvent(e, topic, isUnique))
		})
	}
}

func (b *Bus[T]) removeEvents(topic string, es []Event[T]) {
	if len(es) == 0 {
		b.events.Remove(topic)
		return
	}

	if b.events.Count() == 0 {
		return
	}

	b.events.Upsert(topic, func(oldValue []*event[T], exist bool) []*event[T] {
		events := make([]*event[T], 0, len(oldValue))
		for _, e := range es {
			tag := reflect.ValueOf(e)
			for i := 0; i < len(events); i++ {
				if events[i].tag == tag {
					events = append(events[:i], events[i+1:]...)
					i--
				}
			}
		}
		return events
	})

	b.events.RemoveCb(topic, func(value []*event[T], exists bool) bool {
		return len(value) == 0
	})
}

func (b *Bus[T]) dispatch(topic string, data []T) {
	var (
		removes = make(map[string][]Event[T])
	)

	b.events.GetCb(topic, func(events []*event[T], exists bool) {
		if !exists {
			return
		}
		for _, e := range events {
			if !e.isUnique {
				e.Dispatch(topic, data...)
				continue
			}
			if atomic.CompareAndSwapUint32(&e.hasCalled, 0, 1) {
				e.Dispatch(topic, data...)
				removes[e.topic] = append(removes[e.topic], e.Event)
			}
		}
	})

	if topic != ALL {
		b.events.GetCb(ALL, func(events []*event[T], exists bool) {
			if !exists {
				return
			}
			for _, e := range events {
				if !e.isUnique {
					e.Dispatch(topic, data...)
					continue
				}
				if atomic.CompareAndSwapUint32(&e.hasCalled, 0, 1) {
					e.Dispatch(topic, data...)
					removes[ALL] = append(removes[ALL], e.Event)
				}
			}
		})
	}

	for k, v := range removes {
		b.removeEvents(k, v)
	}
}
