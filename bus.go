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
	topics        cmap.ConcurrentMap[string, *Topic[T]]
}

// New - return a new Bus topicHandlerject
func New[T any]() *Bus[T] {
	return &Bus[T]{
		topics: cmap.New[*Topic[T]](),
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
	go func(oldTopics cmap.ConcurrentMap[string, *Topic[T]]) {
		oldTopics.IterCb(func(_ string, topicHandler *Topic[T]) {
			topicHandler.Clear()
		})
	}(b.topics)
	b.topics = cmap.New[*Topic[T]]()
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

// TopicCount - return the number of topics
func (b *Bus[T]) TopicCount() int {
	return b.topics.Count()
}

// EventCount - return the number of events for a topic
func (b *Bus[T]) EventCount(topic string) int {
	topicHandler, ok := b.topics.Get(topic)
	if !ok {
		return 0
	}
	return topicHandler.Count()
}

// TotalEvents - return the total number of events
func (b *Bus[T]) TotalEvents() int {
	total := 0
	b.topics.IterCb(func(_ string, topicHandler *Topic[T]) {
		total += topicHandler.Count()
	})
	return total
}

// Get - get topic handler
func (b *Bus[T]) Get(topic string) *Topic[T] {
	topicHandler, ok := b.topics.Get(topic)
	if !ok {
		return nil
	}
	return topicHandler
}

func (b *Bus[T]) addEvent(topic string, isUnique bool, e Event[T]) {
	b.topics.Upsert(topic, func(oldValue *Topic[T], exist bool) *Topic[T] {
		if !exist {
			oldValue = &Topic[T]{
				name:   topic,
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

	b.topics.RemoveCb(topic, func(topicHandler *Topic[T], exists bool) bool {
		if !exists {
			return true
		}
		removed = topicHandler.removeEvents(vs)
		return topicHandler.Count() == 0
	})

	if len(removed) > 0 {
		go onStop(topic, removed)
	}
}

func (b *Bus[T]) dispatch(topic string, data []T) {
	b.topics.GetCb(topic, func(topicHandler *Topic[T], exists bool) {
		if !exists {
			return
		}
		topicHandler.Dispatch(data)
	})

	if b.allowAsterisk && topic != ALL {
		b.topics.GetCb(ALL, func(topicHandler *Topic[T], exists bool) {
			if !exists {
				return
			}
			topicHandler.Dispatch(data)
		})
	}
}

func (b *Bus[T]) broadcast(data []T) {
	b.topics.IterCb(func(_ string, topicHandler *Topic[T]) {
		topicHandler.Dispatch(data)
	})
}
