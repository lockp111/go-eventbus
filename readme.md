# go-eventbus

A high-performance event bus model for Go, supporting generics, thread-safe, and suitable for high-concurrency operations.

[简体中文](README-zh.md) | English

## Features

- **Type Safety**: Utilizing Go 1.18+ generics support for type-safe event handling
- **High Performance**: Optimized for concurrent processing, suitable for high-load scenarios
- **Thread Safety**: Support for high-concurrency read/write operations
- **Chainable API**: Convenient API design with chainable method calls
- **Flexible Subscription**: Support for one-time events and persistent subscriptions

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

### Triggering All Events - TriggerAll(msg ...T) *Bus[T]

Send messages to all topics:

```go
// Send a message to all topics
bus.TriggerAll("Broadcast message")
```

### Global Events - Using the ALL Constant

Subscribe to events for all topics:

```go
// Subscribe to all topics
bus.On(eventbus.ALL, &GlobalHandler{})
```

### Querying Statistics

```go
// Get the number of event handlers for a specific topic
count := bus.Count("message")
fmt.Printf("Topic 'message' has %d handlers\n", count)

// Get the total number of events
total := bus.Total()
fmt.Printf("Event bus has %d events in total\n", total)
```

### Clearing All Events - Clean() *Bus[T]

```go
// Clear all event subscriptions
bus.Clean()
```

## Advanced Examples

### Custom Event Handler

```go
type LogEventHandler struct {
    logger *log.Logger
    level  string
}

func (h *LogEventHandler) Dispatch(topic string, messages []string) {
    for _, msg := range messages {
        h.logger.Printf("[%s] %s: %s", h.level, topic, msg)
    }
}

// Create and subscribe
logger := log.New(os.Stdout, "", log.LstdFlags)
bus.On("error", &LogEventHandler{logger, "ERROR"})
bus.On("info", &LogEventHandler{logger, "INFO"})

// Trigger log events
bus.Trigger("error", "System exception")
bus.Trigger("info", "User logged in successfully")
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