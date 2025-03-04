package eventbus

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
)

/**
 * 基准测试部分
 */

var eventsList = []string{"foo", "bar", "baz", "boo"}

func BenchmarkOnTrigger(b *testing.B) {
	o := New[string]()
	n := 0

	fn := &StringEventHandler{&n, ""}
	for _, e := range eventsList {
		o.On(e, fn)
	}

	for i := 0; i < b.N; i++ {
		for _, e := range eventsList {
			o.Trigger(e)
		}
	}
}

// 模拟事件处理器
type benchmarkEvent struct {
	counter *int64
}

func (e *benchmarkEvent) Dispatch(topic string, _ []string) {
	atomic.AddInt64(e.counter, 1)
}

// 基准测试：并发订阅和触发事件
func BenchmarkConcurrentSubscribeAndTrigger(b *testing.B) {
	bus := New[string]()
	var counter int64

	// 预热：订阅大量事件
	for i := 0; i < 1000; i++ {
		topic := fmt.Sprintf("topic-%d", i)
		bus.On(topic, &benchmarkEvent{&counter})
	}

	b.ResetTimer() // 重置计时器，不计算预热时间

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			topic := fmt.Sprintf("topic-%d", rand.Intn(1000))
			bus.Trigger(topic, "test message")
		}
	})

	b.StopTimer() // 停止计时器

	b.ReportMetric(float64(atomic.LoadInt64(&counter)), "dispatches")
}

// 基准测试：频繁订阅和取消订阅
func BenchmarkSubscribeUnsubscribe(b *testing.B) {
	bus := New[string]()
	var counter int64

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		topic := fmt.Sprintf("topic-%d", i%100)
		event := &benchmarkEvent{&counter}
		bus.On(topic, event)
		bus.Off(topic, event)
	}
}

// 基准测试：大量事件的内存使用
func BenchmarkMemoryUsage(b *testing.B) {
	bus := New[string]()
	var counter int64

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		topic := fmt.Sprintf("topic-%d", i)
		bus.On(topic, &benchmarkEvent{&counter})
		if i%100 == 0 {
			runtime.GC() // 强制进行垃圾回收
		}
	}

	b.StopTimer()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	b.ReportMetric(float64(m.Alloc), "bytes_allocated")
}

// 基准测试：并发订阅、触发和取消订阅
func BenchmarkConcurrentOperations(b *testing.B) {
	bus := New[string]()
	var counter int64
	var wg sync.WaitGroup

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		wg.Add(3)
		go func() {
			defer wg.Done()
			topic := fmt.Sprintf("topic-%d", rand.Intn(1000))
			event := &benchmarkEvent{&counter}
			bus.On(topic, event)
		}()
		go func() {
			defer wg.Done()
			topic := fmt.Sprintf("topic-%d", rand.Intn(1000))
			bus.Trigger(topic, "test message")
		}()
		go func() {
			defer wg.Done()
			topic := fmt.Sprintf("topic-%d", rand.Intn(1000))
			event := &benchmarkEvent{&counter}
			bus.Off(topic, event)
		}()
	}

	wg.Wait()
	b.StopTimer()

	b.ReportMetric(float64(atomic.LoadInt64(&counter)), "dispatches")
}
