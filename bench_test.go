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

// 添加OnStop方法以实现Event接口
func (e *benchmarkEvent) OnStop(topic string) {
	// 基准测试处理器不需要特殊处理
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

// 基准测试：高并发读写场景
func BenchmarkHighConcurrentReadWrite(b *testing.B) {
	// 定义不同的并发数进行测试
	concurrencyLevels := []int{100, 1000, 10000, 100000}

	for _, concurrency := range concurrencyLevels {
		b.Run(fmt.Sprintf("Concurrency-%d", concurrency), func(b *testing.B) {
			// 设置并发数
			runtime.GOMAXPROCS(runtime.NumCPU())

			bus := New[string]()
			var counter int64

			// 预热：初始化一些事件
			for i := 0; i < 1000; i++ {
				topic := fmt.Sprintf("init-topic-%d", i)
				bus.On(topic, &benchmarkEvent{&counter})
			}

			b.ResetTimer()

			// 创建一个WaitGroup来等待所有goroutine完成
			var wg sync.WaitGroup
			wg.Add(concurrency)

			// 启动指定数量的goroutine
			for i := 0; i < concurrency; i++ {
				go func(id int) {
					defer wg.Done()

					// 本地随机数生成器，避免锁竞争
					localRand := rand.New(rand.NewSource(rand.Int63() + int64(id)))

					// 每个goroutine执行b.N/concurrency次操作
					iterationsPerGoroutine := max(b.N/concurrency, 1)

					for j := 0; j < iterationsPerGoroutine; j++ {
						// 随机选择操作：读(触发事件)、写(订阅/取消订阅)
						op := localRand.Intn(1000)
						topic := fmt.Sprintf("topic-%d", localRand.Intn(1000))

						if op < 800 { // 80%的概率是读操作(触发事件)
							bus.Trigger(topic, "message")
						} else if op < 900 { // 10%的概率是订阅操作
							event := &benchmarkEvent{&counter}
							bus.On(topic, event)
						} else { // 10%的概率是取消订阅操作
							event := &benchmarkEvent{&counter}
							bus.Off(topic, event) // 然后取消订阅
						}
					}
				}(i)
			}

			// 等待所有goroutine完成
			wg.Wait()
			b.StopTimer()

			// 报告有多少事件被处理
			b.ReportMetric(float64(atomic.LoadInt64(&counter)), "dispatches")

			// 报告事件总数
			b.ReportMetric(float64(bus.Total()), "total_events")

			// 获取内存使用情况
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			b.ReportMetric(float64(m.Alloc)/1024/1024, "memory_MB")
		})
	}
}
