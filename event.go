package eventbus

import (
	"reflect"
)

// ALL - The key use to listen all the topics
const ALL = "*"

// Event interface
type Event interface {
	Dispatch(data any)
}

// event struct
type event struct {
	Event
	topic     string
	tag       reflect.Value
	isUnique  bool
	hasCalled bool
}

func newEvent(e Event, topic string, isUnique bool) *event {
	return &event{e, topic, reflect.ValueOf(e), isUnique, false}
}
