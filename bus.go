package eventbus

import (
	"reflect"
	"sync"
)

// Bus struct
type Bus struct {
	events map[string][]*event
	mux    sync.RWMutex
}

// New - return a new Bus object
func New() *Bus {
	return &Bus{
		events: make(map[string][]*event),
	}
}

// On - register topic event and return error
func (b *Bus) On(topic string, e ...Event) *Bus {
	b.addEvents(topic, false, e)
	return b
}

// Once - register once event and return error
func (b *Bus) Once(topic string, e ...Event) *Bus {
	b.addEvents(topic, true, e)
	return b
}

// Off - remove topic event
func (b *Bus) Off(topic string, e ...Event) *Bus {
	b.removeEvents(topic, e)
	return b
}

// Trigger - dispatch event
func (b *Bus) Trigger(topic string, msg ...interface{}) *Bus {
	if len(msg) != 0 {
		for _, d := range msg {
			b.dispatch(topic, d)
		}
	} else {
		b.dispatch(topic, nil)
	}

	return b
}

func (b *Bus) addEvents(topic string, isUnique bool, es []Event) {
	if len(es) == 0 {
		return
	}

	b.mux.Lock()
	defer b.mux.Unlock()

	for _, e := range es {
		b.events[topic] = append(b.events[topic], newEvent(e, isUnique))
	}
}

func (b *Bus) removeEvents(topic string, es []Event) {
	b.mux.Lock()
	defer b.mux.Unlock()

	if topic == ALL_TOPICS {
		b.events = make(map[string][]*event)
		return
	}

	if len(es) == 0 {
		delete(b.events, topic)
		return
	}

	events := b.events[topic]
	if len(events) == 0 {
		return
	}

	for _, e := range es {
		tag := reflect.ValueOf(e)
		for i := 0; i < len(events); i++ {
			if events[i].tag == tag {
				events = append(events[:i], events[i+1:]...)
				i--
			}
		}
	}

	if len(events) == 0 {
		delete(b.events, topic)
	} else {
		b.events[topic] = events
	}
}

func (b *Bus) dispatch(topic string, data interface{}) {
	b.mux.RLock()
	defer b.mux.RUnlock()

	if topic == ALL_TOPICS {
		for _, es := range b.events {
			for _, e := range es {
				e.dispatch(data)
			}
		}
	} else {
		for _, e := range b.events[topic] {
			e.dispatch(data)
		}

		for _, e := range b.events[ALL_TOPICS] {
			e.dispatch(data)
		}
	}

}
