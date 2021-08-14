package eventbus

var defaultBus = New()

// On - register topic event and return error
func On(topic string, e ...Event) *Bus {
	return defaultBus.On(topic, e...)
}

// Once - register once event and return error
func Once(topic string, e ...Event) *Bus {
	return defaultBus.Once(topic, e...)
}

// Off - remove topic event
func Off(topic string, e ...Event) *Bus {
	return defaultBus.Off(topic, e...)
}

// Clean - clear all events
func Clean() *Bus {
	return defaultBus.Clean()
}

// Trigger - dispatch event
func Trigger(topic string, msg ...interface{}) *Bus {
	return defaultBus.Trigger(topic, msg...)
}
