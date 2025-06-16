# go-eventbus

A high-performance event bus model for Go, supporting generics, thread-safe, and suitable for high-concurrency operations.

[简体中文](README-zh.md) | English

## Features

- **Type Safety**: Utilizing Go 1.18+ generics support for type-safe event handling
- **High Performance**: Optimized for concurrent processing, suitable for high-load scenarios
- **Thread Safety**: Support for high-concurrency read/write operations
- **Chainable API**: Convenient API design with chainable method calls
- **Flexible Subscription**: Support for one-time events and persistent subscriptions
- **Global Events**: Support for subscribing to all topics using the ALL constant
- **Callback Support**: Rich callback mechanisms for event unsubscription
- **Statistics**: Built-in support for event counting and topic statistics

## Installation

### go get
```go
go get -u github.com/lockp111/go-eventbus
```

## Quick Start

Create an event bus instance:

```go
import "github.com/lockp111/go-eventbus"

// Create an event bus that handles string messages
bus := eventbus.New[string]()
```

## Usage Examples

### Basic Usage

#### Creating an Event Handler

```go
// Define an event handler
type MessageHandler struct{}

// Implement the Dispatch method
func (h *MessageHandler) Dispatch(topic string, messages []string) {
    fmt.Printf("Received message for topic %s: %v\n", topic, messages)
}

// Implement the OnStop method
func (h *MessageHandler) OnStop(topic string) {
    fmt.Printf("Handler stopped for topic %s\n", topic)
}

// Subscribe to an event
bus.On("message", &MessageHandler{})

// Trigger an event
bus.Trigger("message", "Hello", "World")
```

### Subscribing to Events - On(topic string, e Event[T]) *Bus[T]

Subscribe to events for a specific topic:

```go
type ReadyHandler struct{}

func (e *ReadyHandler) Dispatch(topic string, data []string) {
    fmt.Printf("Topic %s is ready, data: %v\n", topic, data)
}

func (e *ReadyHandler) OnStop(topic string) {
    fmt.Printf("ReadyHandler stopped for topic %s\n", topic)
}

// Subscribe to a single event
bus.On("ready", &ReadyHandler{})

// Chain subscribe to multiple topics
bus.On("start", &StartHandler{}).On("stop", &StopHandler{})
```

### One-time Subscription - Once(topic string, e Event[T]) *Bus[T]

Subscribe to a one-time event (automatically unsubscribed after triggering once):

```go
type InitHandler struct{}

func (e *InitHandler) Dispatch(topic string, data []string) {
    fmt.Println("System initialization completed")
}

func (e *InitHandler) OnStop(topic string) {
    fmt.Printf("InitHandler stopped for topic %s\n", topic)
}

// Subscribe to a one-time event
bus.Once("init", &InitHandler{})
```

### Unsubscribing - Off(topic string, es ...Event[T]) *Bus[T]

Unsubscribe from specific events:

```go
handler := &ReadyHandler{}
bus.On("ready", handler)

// Unsubscribe from a specific event
bus.Off("ready", handler)

// Unsubscribe from all events under a topic
bus.Off("ready")
```

### Unsubscribing with Callback - OffCb(topic string, cb OffCallback, es ...Event[T]) *Bus[T]

```go
handler := &ReadyHandler{}
bus.On("ready", handler)

// Unsubscribe and get results
bus.OffCb("ready", func(count int, exists bool) {
    fmt.Printf("Removed %d event handlers, topic exists: %v\n", count, exists)
}, handler)
```

### Triggering Events - Trigger(topic string, msg ...T) *Bus[T]

Send messages to a specific topic:

```go
// Send a single message
bus.Trigger("message", "Hello")

// Send multiple messages
bus.Trigger("message", "Hello", "World", "!")
```

### Broadcasting Events - Broadcast(msg ...T) *Bus[T]

Send messages to all topics:

```go
// Send a message to all topics
bus.Broadcast("Broadcast message")
```

### Global Events - Using the ALL Constant

Subscribe to events for all topics:

```go
// Subscribe to all topics
bus.On(eventbus.ALL, &GlobalHandler{})

// Allow asterisk topic for global event handling
bus.AllowAsterisk()
```

### Querying Statistics

```go
// Get the number of event handlers for a specific topic
count := bus.EventCount("message")
fmt.Printf("Topic 'message' has %d handlers\n", count)

// Get the total number of events
total := bus.TotalEvents()
fmt.Printf("Event bus has %d events in total\n", total)

// Get the number of topics
topicCount := bus.TopicCount()
fmt.Printf("Event bus has %d topics\n", topicCount)
```

### Getting Topic Information

```go
// Get topic information
topic := bus.Get("message")
if topic != nil {
    fmt.Printf("Topic has %d handlers\n", topic.Count())
}
```

### Clearing All Events - Clean() *Bus[T]

```go
// Clear all event subscriptions
bus.Clean()
```

## Advanced Examples

### Custom Event Handler with State

```go
type LogEventHandler struct {
    logger *log.Logger
    level  string
    counter int
}

func (h *LogEventHandler) Dispatch(topic string, messages []string) {
    for _, msg := range messages {
        h.counter++
        h.logger.Printf("[%s] %s: %s (count: %d)", h.level, topic, msg, h.counter)
    }
}

func (h *LogEventHandler) OnStop(topic string) {
    h.logger.Printf("[%s] Handler stopped for topic %s\n", h.level, topic)
}

// Create and subscribe
logger := log.New(os.Stdout, "", log.LstdFlags)
bus.On("error", &LogEventHandler{logger, "ERROR", 0})
bus.On("info", &LogEventHandler{logger, "INFO", 0})

// Trigger log events
bus.Trigger("error", "System exception")
bus.Trigger("info", "User logged in successfully")
```

### Concurrent Event Handling

```go
type AtomicHandler struct {
    counter *int32
}

func (h *AtomicHandler) Dispatch(topic string, data []string) {
    atomic.AddInt32(h.counter, 1)
}

func (h *AtomicHandler) OnStop(topic string) {
    // Handle cleanup if needed
}

// Create handler with atomic counter
var counter int32
handler := &AtomicHandler{&counter}

// Subscribe and trigger concurrently
bus.On("concurrent", handler)
for i := 0; i < 100; i++ {
    go bus.Trigger("concurrent", "data")
}
```

## Performance

`go-eventbus` is optimized for high-concurrency scenarios. Here are some benchmark results:

```
BenchmarkOnTrigger-8                    3000000               399 ns/op              54 B/op          1 allocs/op
BenchmarkConcurrentSubscribeAndTrigger-8  100000             17890 ns/op            2156 B/op         21 allocs/op
BenchmarkSubscribeUnsubscribe-8          1000000              1035 ns/op             168 B/op          2 allocs/op
BenchmarkHighConcurrentReadWrite-8         50000             30120 ns/op            3012 B/op         30 allocs/op
```

Run the benchmarks:

```bash
go test -bench=. -benchmem
```

## Concurrency Safety

`go-eventbus` uses thread-safe data structures internally and supports high-concurrency read/write operations without additional synchronization measures.

## Contributing

Issues and PRs welcome!

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Other Languages

- [中文文档](README-zh.md)