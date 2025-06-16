# go-eventbus

适用于Go语言的高性能事件总线模型，支持泛型，线程安全，支持高并发操作。

简体中文 | [English](README.md)

## 特性

- **类型安全**：使用Go 1.18+泛型支持，提供类型安全的事件处理
- **高性能**：优化的并发处理，适用于高负载场景
- **线程安全**：支持高并发读写操作
- **链式调用**：方便的API设计，支持链式调用
- **灵活订阅**：支持一次性事件和持久订阅
- **全局事件**：支持使用ALL常量订阅所有主题
- **回调支持**：丰富的事件取消订阅回调机制
- **统计功能**：内置事件计数和主题统计支持

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

// 实现OnStop方法
func (h *MessageHandler) OnStop(topic string) {
    fmt.Printf("处理器已停止，主题: %s\n", topic)
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

func (e *ReadyHandler) OnStop(topic string) {
    fmt.Printf("ReadyHandler已停止，主题: %s\n", topic)
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

func (e *InitHandler) OnStop(topic string) {
    fmt.Printf("InitHandler已停止，主题: %s\n", topic)
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

### 广播事件 - Broadcast(msg ...T) *Bus[T]

向所有主题发送消息：

```go
// 向所有主题发送消息
bus.Broadcast("广播消息")
```

### 全局事件 - 使用ALL常量

订阅所有主题的事件：

```go
// 订阅所有主题
bus.On(eventbus.ALL, &GlobalHandler{})

// 允许星号主题用于全局事件处理
bus.AllowAsterisk()
```

### 查询统计信息

```go
// 获取特定主题的事件处理器数量
count := bus.EventCount("message")
fmt.Printf("主题'message'有 %d 个处理器\n", count)

// 获取总事件数
total := bus.TotalEvents()
fmt.Printf("事件总线共有 %d 个事件\n", total)

// 获取主题数量
topicCount := bus.TopicCount()
fmt.Printf("事件总线共有 %d 个主题\n", topicCount)
```

### 获取主题信息

```go
// 获取主题信息
topic := bus.Get("message")
if topic != nil {
    fmt.Printf("主题有 %d 个处理器\n", topic.Count())
}
```

### 清除所有事件 - Clean() *Bus[T]

```go
// 清除所有事件订阅
bus.Clean()
```

## 高级示例

### 带状态的自定义事件处理器

```go
type LogEventHandler struct {
    logger *log.Logger
    level  string
    counter int
}

func (h *LogEventHandler) Dispatch(topic string, messages []string) {
    for _, msg := range messages {
        h.counter++
        h.logger.Printf("[%s] %s: %s (计数: %d)", h.level, topic, msg, h.counter)
    }
}

func (h *LogEventHandler) OnStop(topic string) {
    h.logger.Printf("[%s] 处理器已停止，主题: %s\n", h.level, topic)
}

// 创建并订阅
logger := log.New(os.Stdout, "", log.LstdFlags)
bus.On("error", &LogEventHandler{logger, "ERROR", 0})
bus.On("info", &LogEventHandler{logger, "INFO", 0})

// 触发日志事件
bus.Trigger("error", "系统异常")
bus.Trigger("info", "用户登录成功")
```

### 并发事件处理

```go
type AtomicHandler struct {
    counter *int32
}

func (h *AtomicHandler) Dispatch(topic string, data []string) {
    atomic.AddInt32(h.counter, 1)
}

func (h *AtomicHandler) OnStop(topic string) {
    // 处理清理工作（如果需要）
}

// 创建带原子计数器的处理器
var counter int32
handler := &AtomicHandler{&counter}

// 订阅并并发触发
bus.On("concurrent", handler)
for i := 0; i < 100; i++ {
    go bus.Trigger("concurrent", "data")
}
```

## 性能

`go-eventbus` 针对高并发场景进行了优化。以下是一些基准测试结果：

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

## 并发安全性

`go-eventbus` 内部使用线程安全的数据结构，支持高并发读写操作，无需额外的同步措施。

## 贡献

欢迎提交 Issue 和 Pull Request！

## 许可证

本项目采用 MIT 许可证。详见 [LICENSE](LICENSE) 文件。

## 其他语言

- [English Documentation](README.md) 