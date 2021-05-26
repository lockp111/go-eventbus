package eventbus

import (
	"reflect"
	"sync"
)

// ALL_TOPICS - The key use to listen or remove all the topics
const ALL_TOPICS = "*"

// Event interface
type Event interface {
	Dispatch(data interface{})
}

// event struct
type event struct {
	Event
	tag       reflect.Value
	isUnique  bool
	hasCalled bool
	sync.RWMutex
}

func newEvent(e Event, isUnique bool) *event {
	return &event{e, reflect.ValueOf(e), isUnique, false, sync.RWMutex{}}
}

func (e *event) dispatch(data interface{}) {
	e.Lock()
	defer e.Unlock()

	if e.isUnique && e.hasCalled {
		return
	}

	e.Dispatch(data)

	if e.isUnique {
		e.hasCalled = true
	}
}
