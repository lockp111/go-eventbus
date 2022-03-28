package eventbus

import (
	"reflect"
	"sync"
)

// Bus struct
type Bus struct {
	mux    sync.Mutex
	events map[string][]*event
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

// Clean - clear all events
func (b *Bus) Clean() *Bus {
	b.mux.Lock()
	defer b.mux.Unlock()
	b.events = make(map[string][]*event)
	return b
}

// Trigger - dispatch event
func (b *Bus) Trigger(topic string, msg ...any) *Bus {
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
		b.events[topic] = append(b.events[topic], newEvent(e, topic, isUnique))
	}
}

func (b *Bus) removeEvents(topic string, es []Event) {
	b.mux.Lock()
	defer b.mux.Unlock()

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
		return
	}

	b.events[topic] = events
}

func (b *Bus) getEvents(topic string) []*event {
	b.mux.Lock()
	defer b.mux.Unlock()

	events := make([]*event, 0, len(b.events[topic])+len(b.events[ALL]))
	for _, e := range b.events[topic] {
		if e.isUnique {
			if e.hasCalled {
				continue
			}
			e.hasCalled = true
		}
		events = append(events, e)
	}

	if topic != ALL {
		for _, e := range b.events[ALL] {
			if e.isUnique {
				if e.hasCalled {
					continue
				}
				e.hasCalled = true
			}
			events = append(events, e)
		}
	}
	return events
}

func (b *Bus) dispatch(topic string, data any) {
	var (
		events  = b.getEvents(topic)
		removes = make(map[string][]Event)
	)

	for _, e := range events {
		e.Dispatch(data)
		if e.isUnique && e.hasCalled {
			removes[e.topic] = append(removes[e.topic], e.Event)
		}
	}

	for k, v := range removes {
		b.removeEvents(k, v)
	}
}
