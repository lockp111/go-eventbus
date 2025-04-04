package eventbus

import (
	"reflect"
)

// Event interface
type Event[T any] interface {
	Dispatch(topic string, data []T)
	OnStop(topic string)
}

// event struct
type event[T any] struct {
	Event[T]
	tag       reflect.Value
	isUnique  bool
	hasCalled uint32
}

func newEvent[T any](e Event[T], isUnique bool) *event[T] {
	return &event[T]{e, reflect.ValueOf(e), isUnique, 0}
}
