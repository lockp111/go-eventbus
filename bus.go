package eventbus

import (
	"reflect"

	"github.com/lockp111/go-cmap"
)

// ALL - The key use to listen all the topics
const ALL = "*"

// Bus struct
type Bus[T any] struct {
	allowAsterisk bool
	topics        cmap.ConcurrentMap[string, *Observer[T]]
}

// New - return a new Bus object
func New[T any]() *Bus[T] {
	return &Bus[T]{
		topics: cmap.New[*Observer[T]](),
	}
}

// AllowAsterisk - allow asterisk topic
func (b *Bus[T]) AllowAsterisk() *Bus[T] {
	b.allowAsterisk = true
	return b
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

// Clean - clear all events
func (b *Bus[T]) Clean() *Bus[T] {
	go func(oldTopics cmap.ConcurrentMap[string, *Observer[T]]) {
		oldTopics.IterCb(func(_ string, ob *Observer[T]) {
			ob.Clear()
		})
	}(b.topics)
	b.topics = cmap.New[*Observer[T]]()
	return b
}

// Trigger - dispatch event
func (b *Bus[T]) Trigger(topic string, msg ...T) *Bus[T] {
	b.dispatch(topic, msg)
	return b
}

// Broadcast - broadcast all topics
func (b *Bus[T]) Broadcast(msg ...T) *Bus[T] {
	b.broadcast(msg)
	return b
}

// Count - topic count events
func (b *Bus[T]) Count(topic string) int {
	ob, ok := b.topics.Get(topic)
	if !ok {
		return 0
	}
	return ob.Count()
}

// Total - total events
func (b *Bus[T]) Total() int {
	total := 0
	b.topics.IterCb(func(_ string, ob *Observer[T]) {
		total += ob.Count()
	})
	return total
}

// Get - get topic observer
func (b *Bus[T]) Get(topic string) *Observer[T] {
	ob, ok := b.topics.Get(topic)
	if !ok {
		return nil
	}
	return ob
}

func (b *Bus[T]) addEvent(topic string, isUnique bool, e Event[T]) {
	b.topics.Upsert(topic, func(oldValue *Observer[T], exist bool) *Observer[T] {
		if !exist {
			oldValue = &Observer[T]{
				topic:  topic,
				events: make([]*event[T], 0),
			}
		}

		oldValue.addEvent(e, isUnique)
		return oldValue
	})
}

func (b *Bus[T]) removeEvents(topic string, es []Event[T]) {
	var (
		removed []*event[T]
		vs      = make([]reflect.Value, 0, len(es))
	)
	for _, e := range es {
		vs = append(vs, reflect.ValueOf(e))
	}

	b.topics.RemoveCb(topic, func(ob *Observer[T], exists bool) bool {
		if !exists {
			return true
		}
		removed = ob.removeEvents(vs)
		return ob.Count() == 0
	})

	if len(removed) > 0 {
		go onStop(topic, removed)
	}
}

func (b *Bus[T]) dispatch(topic string, data []T) {
	b.topics.GetCb(topic, func(ob *Observer[T], exists bool) {
		if !exists {
			return
		}
		ob.Distpatch(data)
	})

	if b.allowAsterisk && topic != ALL {
		b.topics.GetCb(ALL, func(ob *Observer[T], exists bool) {
			if !exists {
				return
			}
			ob.Distpatch(data)
		})
	}
}

func (b *Bus[T]) broadcast(data []T) {
	b.topics.IterCb(func(_ string, ob *Observer[T]) {
		ob.Distpatch(data)
	})
}
