package eventbus

import (
	"reflect"
	"sync"
)

// ALL_TOPICS - The key use to listen or remove all the topics
const ALL_TOPICS = "*"

// Event interface
type Event interface {
	Dispatch(msg interface{})
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

func (e *event) isCall() bool {
	e.RLock()
	defer e.RUnlock()

	if e.isUnique && e.hasCalled {
		return false
	}

	return true
}

func (e *event) dispatch(msg interface{}) {
	if !e.isCall() {
		return
	}

	e.Dispatch(msg)

	if e.isUnique {
		e.Lock()
		e.hasCalled = true
		e.Unlock()
	}
}
