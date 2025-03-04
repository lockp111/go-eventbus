package eventbus

import (
	"reflect"
	"sync/atomic"

	"github.com/lockp111/go-cmap"
)

type OffCallback func(count int, exists bool)

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
func (b *Bus[T]) On(topic string, e Event[T]) *Bus[T] {
	b.addEvent(topic, false, e)
	return b
}

// Once - register once event and return error
func (b *Bus[T]) Once(topic string, e Event[T]) *Bus[T] {
	b.addEvent(topic, true, e)
	return b
}

// Off - remove topic event
func (b *Bus[T]) Off(topic string, es ...Event[T]) *Bus[T] {
	b.removeEvents(topic, es)
	return b
}

// OffCb - remove topic event and callback
func (b *Bus[T]) OffCb(topic string, cb OffCallback, es ...Event[T]) *Bus[T] {
	b.removeEventsCb(topic, es, cb)
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

// TriggerAll - dispatch all topics
func (b *Bus[T]) TriggerAll(msg ...T) *Bus[T] {
	b.dispatchAll(msg)
	return b
}

// Count - topic count events
func (b *Bus[T]) Count(topic string) int {
	es, _ := b.events.Get(topic)
	return len(es)
}

// Total - total events
func (b *Bus[T]) Total() int {
	return b.events.Count()
}

func (b *Bus[T]) addEvent(topic string, isUnique bool, e Event[T]) {
	b.events.Upsert(topic, func(oldValue []*event[T], _ bool) []*event[T] {
		return append(oldValue, newEvent(e, topic, isUnique))
	})
}

func (b *Bus[T]) removeEvents(topic string, es []Event[T]) {
	b.removeEventsCb(topic, es)
}

func (b *Bus[T]) removeEventsCb(topic string, es []Event[T], cb ...OffCallback) {
	if len(es) == 0 {
		b.events.RemoveCb(topic, func(_ []*event[T], exists bool) bool {
			for _, callback := range cb {
				callback(0, exists)
			}
			return true
		})
		return
	}

	b.events.Upsert(topic, func(oldValue []*event[T], exist bool) []*event[T] {
		if !exist || len(oldValue) == 0 {
			return []*event[T]{}
		}
		for _, e := range es {
			tag := reflect.ValueOf(e)
			for i, v := range oldValue {
				if v.tag == tag {
					oldValue = append(oldValue[:i], oldValue[i+1:]...)
				}
			}
		}
		return oldValue
	})

	b.events.RemoveCb(topic, func(value []*event[T], exists bool) bool {
		count := len(value)
		for _, callback := range cb {
			callback(count, exists)
		}
		return count == 0
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
		dispatch(topic, events, removes, data)
	})

	if topic != ALL {
		b.events.GetCb(ALL, func(events []*event[T], exists bool) {
			if !exists {
				return
			}
			dispatch(topic, events, removes, data)
		})
	}

	for k, v := range removes {
		b.removeEvents(k, v)
	}
}

func (b *Bus[T]) dispatchAll(data []T) {
	var (
		removes = make(map[string][]Event[T])
	)

	b.events.IterCb(func(topic string, events []*event[T]) {
		dispatch(topic, events, removes, data)
	})

	for k, v := range removes {
		b.removeEvents(k, v)
	}
}

func dispatch[T any](topic string, events []*event[T], removes map[string][]Event[T], data []T) {
	for _, e := range events {
		if !e.isUnique {
			e.Dispatch(topic, data)
			continue
		}
		if atomic.CompareAndSwapUint32(&e.hasCalled, 0, 1) {
			e.Dispatch(topic, data)
			removes[e.topic] = append(removes[e.topic], e.Event)
		}
	}
}
