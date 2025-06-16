package eventbus

import (
	"reflect"
	"slices"
	"sync/atomic"
)

// Topic represents a topic and its associated event handlers
type Topic[T any] struct {
	name          string
	events        []*event[T]
	allowAsterisk *bool
	asterisk      *Topic[T]
}

func newTopic[T any](name string, allowAsterisk *bool, asterisk *Topic[T]) *Topic[T] {
	topic := &Topic[T]{
		name:   name,
		events: make([]*event[T], 0),
	}
	if name != ALL {
		topic.allowAsterisk = allowAsterisk
		topic.asterisk = asterisk
	}
	return topic
}

func (t *Topic[T]) Dispatch(data ...T) {
	var (
		removed []*event[T]
		removes = t.dispatch(data)
	)

	if t.name != ALL && *t.allowAsterisk {
		removes = append(removes, t.asterisk.dispatch(data)...)
	}

	if len(removes) > 0 {
		removed = t.removeEvents(removes)
	}

	if len(removed) > 0 {
		go onStop(t.name, removed)
	}
}

func (t *Topic[T]) Count() int {
	return len(t.events)
}

func (t *Topic[T]) Clear() {
	removed := t.removeEvents(nil)
	onStop(t.name, removed)
}

func (t *Topic[T]) addEvent(e Event[T], isUnique bool) {
	t.events = append(t.events, newEvent(e, isUnique))
}

func (t *Topic[T]) removeEvents(es []reflect.Value) []*event[T] {
	if len(es) == 0 {
		removed := t.events
		t.events = nil
		return removed
	}

	newEvents := make([]*event[T], 0, len(t.events))
	removed := make([]*event[T], 0, len(es))
	for _, e := range t.events {
		if slices.Contains(es, e.tag) {
			removed = append(removed, e)
			continue
		}
		newEvents = append(newEvents, e)
	}
	t.events = newEvents
	return removed
}

func (t *Topic[T]) dispatch(data []T) []reflect.Value {
	var removes = make([]reflect.Value, 0, 1024)
	for _, e := range t.events {
		if !e.isUnique {
			e.Dispatch(t.name, data)
			continue
		}
		if atomic.CompareAndSwapUint32(&e.hasCalled, 0, 1) {
			e.Dispatch(t.name, data)
			removes = append(removes, e.tag)
		}
	}

	return removes
}

func onStop[T any](topic string, es []*event[T]) {
	for _, e := range es {
		e.OnStop(topic)
	}
}
