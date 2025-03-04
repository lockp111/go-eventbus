# go-eventbus

适用于Go语言的高性能事件总线模型，支持泛型，线程安全，支持高并发操作。

简体中文 | [English](README.md)

## 特性

- **类型安全**：使用Go 1.18+泛型支持，提供类型安全的事件处理
- **高性能**：优化的并发处理，适用于高负载场景
- **线程安全**：支持高并发读写操作
- **链式调用**：方便的API设计，支持链式调用
- **灵活订阅**：支持一次性事件和持久订阅

## 安装

### go get
```go
go get -u github.com/lockp111/go-eventbus
```

## 快速开始

创建一个事件总线实例：

```go
import "github.com/lockp111/go-eventbus"

// 创建一个处理string类型消息的事件总线
bus := eventbus.New[string]()
```

## 使用示例

### 基本用法

#### 创建事件处理器

```go
// 定义一个事件处理器
type MessageHandler struct{}

// 实现Dispatch方法
func (h *MessageHandler) Dispatch(topic string, messages []string) {
    fmt.Printf("收到主题 %s 的消息: %v\n", topic, messages)
}

// 订阅事件
bus.On("message", &MessageHandler{})

// 触发事件
bus.Trigger("message", "Hello", "World")
```

### 订阅事件 - On(topic string, e Event[T]) *Bus[T]

订阅指定主题的事件：

```go
type ReadyHandler struct{}

func (e *ReadyHandler) Dispatch(topic string, data []string) {
    fmt.Printf("主题 %s 已就绪，数据: %v\n", topic, data)
}

// 订阅单个事件
bus.On("ready", &ReadyHandler{})

// 链式订阅多个主题
bus.On("start", &StartHandler{}).On("stop", &StopHandler{})
```

### 一次性订阅 - Once(topic string, e Event[T]) *Bus[T]

订阅一次性事件（触发一次后自动取消订阅）：

```go
type InitHandler struct{}

func (e *InitHandler) Dispatch(topic string, data []string) {
    fmt.Println("系统初始化完成")
}

// 订阅一次性事件
bus.Once("init", &InitHandler{})
```

### 取消订阅 - Off(topic string, es ...Event[T]) *Bus[T]

取消特定事件的订阅：

```go
handler := &ReadyHandler{}
bus.On("ready", handler)

// 取消特定事件的订阅
bus.Off("ready", handler)

// 取消主题下的所有订阅
bus.Off("ready")
```

### 带回调的取消订阅 - OffCb(topic string, cb OffCallback, es ...Event[T]) *Bus[T]

```go
handler := &ReadyHandler{}
bus.On("ready", handler)

// 取消订阅并获取结果
bus.OffCb("ready", func(count int, exists bool) {
    fmt.Printf("已移除 %d 个事件处理器，主题是否存在: %v\n", count, exists)
}, handler)
```

### 触发事件 - Trigger(topic string, msg ...T) *Bus[T]

向特定主题发送消息：

```go
// 发送单个消息
bus.Trigger("message", "Hello")

// 发送多个消息
bus.Trigger("message", "Hello", "World", "!")
```

### 触发所有事件 - TriggerAll(msg ...T) *Bus[T]

向所有主题发送消息：

```go
// 向所有主题发送消息
bus.TriggerAll("Broadcast message")
```

### 全局事件 - 使用ALL常量

订阅所有主题的事件：

```go
// 订阅所有主题
bus.On(eventbus.ALL, &GlobalHandler{})
```

### 查询统计信息

```go
// 获取特定主题的事件处理器数量
count := bus.Count("message")
fmt.Printf("主题'message'有 %d 个处理器\n", count)

// 获取总事件数
total := bus.Total()
fmt.Printf("事件总线共有 %d 个事件\n", total)
```

### 清除所有事件 - Clean() *Bus[T]

```go
// 清除所有事件订阅
bus.Clean()
```

## 高级示例

### 自定义事件处理器

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

// 创建并订阅
logger := log.New(os.Stdout, "", log.LstdFlags)
bus.On("error", &LogEventHandler{logger, "ERROR"})
bus.On("info", &LogEventHandler{logger, "INFO"})

// 触发日志事件
bus.Trigger("error", "系统异常")
bus.Trigger("info", "用户登录成功")
```

## 性能

`go-eventbus`针对高并发场景进行了优化，以下是一些基准测试结果：

```
BenchmarkOnTrigger-8                    3000000               399 ns/op              54 B/op          1 allocs/op
BenchmarkConcurrentSubscribeAndTrigger-8  100000             17890 ns/op            2156 B/op         21 allocs/op
BenchmarkSubscribeUnsubscribe-8          1000000              1035 ns/op             168 B/op          2 allocs/op
BenchmarkHighConcurrentReadWrite-8         50000             30120 ns/op            3012 B/op         30 allocs/op
```

运行基准测试：

```bash
go test -bench=. -benchmem
```

## 并发安全

`go-eventbus`内部使用了线程安全的数据结构，支持高并发环境下的读写操作，无需额外的同步措施。

## 贡献

欢迎提交问题和PR！

## 许可证

本项目采用MIT许可证。详情请查看[LICENSE](LICENSE)文件。 