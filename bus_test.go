package eventbus

import (
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
)

type N struct {
	i *int
	s string
}

func (n *N) Dispatch(topic string, data []string) {
	*n.i++
	for _, s := range data {
		n.s = s
		log.Println(s)
	}
}

type input struct {
	s string
	b bool
}

type FN struct {
	i *int
}

func (fn *FN) Dispatch(topic string, data []input) {
	*fn.i++
	if len(data) == 0 {
		return
	}
	if data[0].b != true || data[0].s != "bar" {
		log.Fatal("The arguments must be correctly passed to the callback")
	}
}

func TestOn(t *testing.T) {
	o := New[string]()
	n := 0

	o.On("foo", &N{&n, ""}).On("bar", &N{&n, ""}).On("foo", &N{&n, ""})
	o.Trigger("foo").Trigger("foo").Trigger("bar")

	if n != 5 {
		t.Errorf("The counter is %d instead of being %d", n, 5)
	}
}

func TestOnAll(t *testing.T) {
	o := New[string]()
	n := 0

	onAll := &N{&n, ""}

	o.On(ALL, onAll)

	o.Trigger("foo", "foo").Trigger("bar", "bar")

	o.Off(ALL, onAll)

	o.Trigger("bar", "bar").Trigger("foo", "bar")

	if onAll.s != "bar" {
		t.Errorf("The last event name triggered is %s instead of being %s", onAll.s, "bar")
	}

	if n != 2 {
		t.Errorf("The counter is %d instead of being %d", n, 2)
	}

}

func TestClean(t *testing.T) {
	o := New[string]()
	n := 0

	fn := &N{&n, ""}
	o.On("foo", fn)

	o.On("bar", fn)

	o.Clean()

	o.Trigger("foo").Trigger("bar").Trigger("foo bar")

	if n != 0 {
		t.Errorf("The counter is %d instead of being %d", n, 0)
	}
}

func TestOff(t *testing.T) {
	o := New[string]()
	n := 0

	onFoo1 := &N{&n, ""}

	onFoo2 := &N{&n, ""}

	o.On("foo", onFoo1).On("foo", onFoo2)
	o.Trigger("foo", "test1")
	if onFoo1.s != "test1" {
		t.Fail()
	}
	if onFoo2.s != "test1" {
		t.Fail()
	}

	o.Off("foo", onFoo1).Off("foo", onFoo2).On("foo", onFoo1)
	o.Trigger("foo", "test2")
	if onFoo1.s != "test2" {
		t.Fail()
	}
	if onFoo2.s == "test2" {
		t.Fail()
	}

	o.On("foo", onFoo2).Off("foo", onFoo1)
	o.Trigger("foo", "test3")
	if onFoo1.s == "test3" {
		t.Fail()
	}
	if onFoo2.s != "test3" {
		t.Fail()
	}

	if n != 4 {
		t.Errorf("The counter is %d instead of being %d", n, 3)
	}

}

func TestRace(t *testing.T) {
	o := New[string]()
	n := 0

	asyncTask := func(wg *sync.WaitGroup) {
		o.Trigger("foo")
		wg.Done()
	}
	var wg sync.WaitGroup

	wg.Add(5)

	fn := &N{&n, ""}
	o.On("foo", fn)

	go asyncTask(&wg)
	go asyncTask(&wg)
	go asyncTask(&wg)
	go asyncTask(&wg)
	go asyncTask(&wg)

	wg.Wait()

	if n != 5 {
		t.Errorf("The counter is %d instead of being %d", n, 5)
	}

}

func TestOne(t *testing.T) {
	o := New[string]()
	n := 0

	onFoo := &N{&n, ""}

	o.Once("foo", onFoo)

	o.Trigger("foo").Trigger("foo").Trigger("foo")

	if n != 1 {
		t.Errorf("The counter is %d instead of being %d", n, 1)
	}

}

func TestArguments(t *testing.T) {
	o := New[input]()
	n := 0
	fn := &FN{&n}
	o.On("foo", fn)

	o.Trigger("foo", input{"bar", true})

	if n != 1 {
		t.Errorf("The counter is %d instead of being %d", n, 1)
	}
}

func TestTrigger(t *testing.T) {
	o := New[interface{}]()
	// the trigger without any listener should not throw errors
	o.Trigger("foo")
}

/**
 * Speed Benchmarks
 */

var eventsList = []string{"foo", "bar", "baz", "boo"}

func BenchmarkOnTrigger(b *testing.B) {
	o := New[string]()
	n := 0

	fn := &N{&n, ""}
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
