package eventbus

import (
	"reflect"
	"slices"
	"sync/atomic"
)

// Observer
type Observer[T any] struct {
	topic  string
	events []*event[T]
}

func (o *Observer[T]) Distpatch(data []T) {
	var (
		removed []*event[T]
		removes = o.dispatch(data)
	)

	if len(removes) > 0 {
		removed = o.removeEvents(removes)
	}

	if len(removed) > 0 {
		go onStop(o.topic, removed)
	}
}

func (o *Observer[T]) Count() int {
	return len(o.events)
}

func (o *Observer[T]) Clear() {
	removed := o.removeEvents(nil)
	onStop(o.topic, removed)
}

func (o *Observer[T]) addEvent(e Event[T], isUnique bool) {
	o.events = append(o.events, newEvent(e, isUnique))
}

func (o *Observer[T]) removeEvents(es []reflect.Value) []*event[T] {
	if len(es) == 0 {
		removed := o.events
		o.events = nil
		return removed
	}

	newEvents := make([]*event[T], 0, len(o.events))
	removed := make([]*event[T], 0, len(es))
	for _, e := range o.events {
		if slices.Contains(es, e.tag) {
			removed = append(removed, e)
			continue
		}
		newEvents = append(newEvents, e)
	}
	o.events = newEvents
	return removed
}

func (o *Observer[T]) dispatch(data []T) []reflect.Value {
	var removes = make([]reflect.Value, 0)
	for _, e := range o.events {
		if !e.isUnique {
			e.Dispatch(o.topic, data)
			continue
		}
		if atomic.CompareAndSwapUint32(&e.hasCalled, 0, 1) {
			e.Dispatch(o.topic, data)
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
