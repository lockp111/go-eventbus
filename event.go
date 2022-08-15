package eventbus

import (
	"reflect"
)

// ALL - The key use to listen all the topics
const ALL = "*"

// Event interface
type Event[T any] interface {
	Dispatch(data ...T)
}

// event struct
type event[T any] struct {
	Event[T]
	topic     string
	tag       reflect.Value
	isUnique  bool
	hasCalled bool
}

func newEvent[T any](e Event[T], topic string, isUnique bool) *event[T] {
	return &event[T]{e, topic, reflect.ValueOf(e), isUnique, false}
}
