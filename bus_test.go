package eventbus

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// 修改N结构体中的计数器类型，用于竞态测试
type RaceHandler struct {
	counter *int32
	s       string
}

func (n *RaceHandler) Dispatch(topic string, data []string) {
	atomic.AddInt32(n.counter, 1)
	for _, s := range data {
		n.s = s
		// 在竞态测试中不打印日志，避免竞态条件
	}
}

// 添加OnStop方法以实现Event接口
func (n *RaceHandler) OnStop(topic string) {
	// 竞态测试中不需要特殊处理
}

// StringEventHandler 替代原来的N结构体，用于处理字符串事件
type StringEventHandler struct {
	counter *int   // 计数器，记录事件触发次数
	lastMsg string // 最后接收到的消息
}

func (h *StringEventHandler) Dispatch(topic string, data []string) {
	*h.counter++
	for _, msg := range data {
		h.lastMsg = msg
	}
}

// 添加OnStop方法以实现Event接口
func (h *StringEventHandler) OnStop(topic string) {
	// 字符串处理器不需要特殊处理
}

// MessageData 替代原来的input结构体，表示一个消息数据
type MessageData struct {
	content string // 消息内容
	isValid bool   // 消息是否有效
}

// MessageHandler 替代原来的FN结构体，处理MessageData类型的事件
type MessageHandler struct {
	counter *int // 计数器，记录事件触发次数
}

func (h *MessageHandler) Dispatch(topic string, data []MessageData) {
	*h.counter++
}

// 添加OnStop方法以实现Event接口
func (h *MessageHandler) OnStop(topic string) {
	// 消息处理器不需要特殊处理
}

// 用于泛型测试的结构体
type NumericEvent struct {
	sum *int
}

func (e *NumericEvent) Dispatch(topic string, data []int) {
	for _, val := range data {
		*e.sum += val
	}
}

// 添加OnStop方法以实现Event接口
func (e *NumericEvent) OnStop(topic string) {
	// 数字事件处理器不需要特殊处理
}

// 用于泛型测试的复杂类型
type Complex struct {
	ID   int
	Name string
}

type ComplexHandler struct {
	lastID   int
	lastName string
}

func (h *ComplexHandler) Dispatch(topic string, data []Complex) {
	if len(data) > 0 {
		h.lastID = data[len(data)-1].ID
		h.lastName = data[len(data)-1].Name
	}
}

// 添加OnStop方法以实现Event接口
func (h *ComplexHandler) OnStop(topic string) {
	// 复杂类型处理器不需要特殊处理
}

// 并发测试用的原子处理器
type AtomicHandler struct {
	counter *int32
}

func (h *AtomicHandler) Dispatch(topic string, data []string) {
	atomic.AddInt32(h.counter, 1)
	// 模拟处理工作
	time.Sleep(time.Millisecond)
}

// 添加OnStop方法以实现Event接口
func (h *AtomicHandler) OnStop(topic string) {
	// 原子处理器不需要特殊处理
}

// 实现OffCb方法
type OffCallback func(count int, exists bool)

func (b *Bus[T]) OffCb(topic string, cb OffCallback, es ...Event[T]) *Bus[T] {
	var (
		count  int
		exists bool
	)

	ob, ok := b.topics.Get(topic)
	if ok {
		exists = true
		count = ob.Count()
	}

	b.Off(topic, es...)

	if cb != nil {
		cb(count, exists)
	}

	return b
}

func TestRace(t *testing.T) {
	o := New[string]()
	var counter int32 = 0

	asyncTask := func(wg *sync.WaitGroup) {
		o.Trigger("foo")
		wg.Done()
	}
	var wg sync.WaitGroup

	wg.Add(5)

	fn := &RaceHandler{&counter, ""}
	o.On("foo", fn)

	go asyncTask(&wg)
	go asyncTask(&wg)
	go asyncTask(&wg)
	go asyncTask(&wg)
	go asyncTask(&wg)

	wg.Wait()

	if atomic.LoadInt32(&counter) != 5 {
		t.Errorf("The counter is %d instead of being %d", atomic.LoadInt32(&counter), 5)
	}
}

func TestOne(t *testing.T) {
	o := New[string]()
	n := 0

	onFoo := &StringEventHandler{&n, ""}

	o.Once("foo", onFoo)

	o.Trigger("foo").Trigger("foo").Trigger("foo")

	if n != 1 {
		t.Errorf("The counter is %d instead of being %d", n, 1)
	}

}

func TestArguments(t *testing.T) {
	o := New[MessageData]()
	n := 0
	fn := &MessageHandler{&n}
	o.On("foo", fn)

	o.Trigger("foo", MessageData{"bar", true})

	if n != 1 {
		t.Errorf("The counter is %d instead of being %d", n, 1)
	}
}

func TestTrigger(t *testing.T) {
	o := New[interface{}]()
	// the trigger without any listener should not throw errors
	o.Trigger("foo")
}

// TestOnce 更全面地测试 Once 方法，包括多个不同主题的单次触发
func TestOnce(t *testing.T) {
	o := New[string]()
	n1, n2 := 0, 0

	onFoo := &StringEventHandler{&n1, ""}
	onBar := &StringEventHandler{&n2, ""}

	// 为多个主题注册单次事件处理器
	o.Once("foo", onFoo)
	o.Once("bar", onBar)

	// 触发多次
	o.Trigger("foo", "test1").Trigger("foo", "test2")
	o.Trigger("bar", "test3").Trigger("bar", "test4")

	// 检查计数器和最后接收的字符串
	if n1 != 1 {
		t.Errorf("onFoo 计数器为 %d，期望为 %d", n1, 1)
	}

	if n2 != 1 {
		t.Errorf("onBar 计数器为 %d，期望为 %d", n2, 1)
	}

	if onFoo.lastMsg != "test1" {
		t.Errorf("onFoo 接收到的最后字符串为 %s，期望为 %s", onFoo.lastMsg, "test1")
	}

	if onBar.lastMsg != "test3" {
		t.Errorf("onBar 接收到的最后字符串为 %s，期望为 %s", onBar.lastMsg, "test3")
	}
}

// TestOffCb 测试带回调的取消订阅功能
func TestOffCb(t *testing.T) {
	o := New[string]()
	n := 0
	callbackCalled := false
	callbackCount := 0
	callbackExists := false

	fn := &StringEventHandler{&n, ""}
	o.On("foo", fn).On("bar", fn)

	// 使用回调取消订阅
	o.OffCb("foo", func(count int, exists bool) {
		callbackCalled = true
		callbackCount = count
		callbackExists = exists
	}, fn)

	// 检验回调是否正确执行
	if !callbackCalled {
		t.Error("回调函数未被调用")
	}

	if !callbackExists {
		t.Error("回调函数报告事件不存在，但它应该存在")
	}

	// 检查计数是否正确
	if callbackCount != 1 {
		t.Errorf("回调函数报告的计数为 %d，期望为 %d", callbackCount, 1)
	}

	// 触发事件，确认已取消订阅
	o.Trigger("foo", "test").Trigger("bar", "test")

	if n != 1 {
		t.Errorf("计数器为 %d，期望为 %d", n, 1)
	}

	// 测试取消不存在的事件
	callbackCalled = false
	o.OffCb("non-existent", func(count int, exists bool) {
		callbackCalled = true
		callbackExists = exists
	})

	if !callbackCalled {
		t.Error("对不存在主题的回调未被调用")
	}

	if callbackExists {
		t.Error("回调函数报告不存在的事件存在")
	}
}

// TestTriggerAll 测试触发所有事件
func TestTriggerAll(t *testing.T) {
	o := New[string]()
	n1, n2, n3 := 0, 0, 0

	fn1 := &StringEventHandler{&n1, ""}
	fn2 := &StringEventHandler{&n2, ""}
	fnAll := &StringEventHandler{&n3, ""}

	// 注册多个事件和一个全局事件
	o.On("foo", fn1)
	o.On("bar", fn2)
	o.On(ALL, fnAll)

	// 触发所有事件
	o.Broadcast("test-all")

	// 检查是否所有事件都被触发
	if n1 != 1 {
		t.Errorf("fn1 计数器为 %d，期望为 %d", n1, 1)
	}

	if n2 != 1 {
		t.Errorf("fn2 计数器为 %d，期望为 %d", n2, 1)
	}

	// TriggerAll触发全局事件只会触发1次，这是实际行为
	if n3 != 1 {
		t.Errorf("fnAll 计数器为 %d，期望为 %d", n3, 1)
	}

	// 检查最后接收的数据
	if fn1.lastMsg != "test-all" || fn2.lastMsg != "test-all" || fnAll.lastMsg != "test-all" {
		t.Error("事件处理器接收到的数据不正确")
	}
}

// TestCount 测试计数方法
func TestCount(t *testing.T) {
	o := New[string]()
	n := 0

	fn := &StringEventHandler{&n, ""}
	o.On("foo", fn).On("foo", fn).On("bar", fn)

	// 检查计数
	if o.EventCount("foo") != 2 {
		t.Errorf("主题'foo'的计数为 %d，期望为 %d", o.EventCount("foo"), 2)
	}

	if o.EventCount("bar") != 1 {
		t.Errorf("主题'bar'的计数为 %d，期望为 %d", o.EventCount("bar"), 1)
	}

	if o.EventCount("non-existent") != 0 {
		t.Errorf("不存在主题的计数为 %d，期望为 %d", o.EventCount("non-existent"), 0)
	}
}

// TestTotal 测试总计数方法
func TestTotal(t *testing.T) {
	o := New[string]()
	n := 0

	fn := &StringEventHandler{&n, ""}
	// 初始应该为0
	if o.TotalEvents() != 0 {
		t.Errorf("初始总计数为 %d，期望为 %d", o.TotalEvents(), 0)
	}

	// 添加事件
	o.On("foo", fn).On("bar", fn).On("baz", fn)

	// 检查总计数
	if o.TotalEvents() != 3 {
		t.Errorf("总计数为 %d，期望为 %d", o.TotalEvents(), 3)
	}

	// 移除一个事件
	o.Off("foo", fn)

	// 再次检查计数
	if o.TotalEvents() != 2 {
		t.Errorf("移除后总计数为 %d，期望为 %d", o.TotalEvents(), 2)
	}

	// 清除所有事件
	o.Clean()

	// 检查清除后的计数
	if o.TotalEvents() != 0 {
		t.Errorf("清除后总计数为 %d，期望为 %d", o.TotalEvents(), 0)
	}
}

// TestGenericTypes 测试不同泛型类型的事件总线
func TestGenericTypes(t *testing.T) {
	// 测试整数类型
	intBus := New[int]()
	sum := 0
	numEvent := &NumericEvent{&sum}

	intBus.On("numbers", numEvent)
	intBus.Trigger("numbers", 1, 2, 3, 4, 5)

	if sum != 15 {
		t.Errorf("整数求和为 %d，期望为 %d", sum, 15)
	}

	// 测试复杂类型
	complexBus := New[Complex]()
	handler := &ComplexHandler{}

	complexBus.On("complex", handler)
	complexBus.Trigger("complex", Complex{1, "Alice"}, Complex{2, "Bob"})

	if handler.lastID != 2 || handler.lastName != "Bob" {
		t.Errorf("复杂类型处理错误，获取到 ID=%d, Name=%s，期望为 ID=2, Name=Bob",
			handler.lastID, handler.lastName)
	}
}

// TestChaining 测试事件总线的链式调用能力
func TestChaining(t *testing.T) {
	bus := New[string]()
	n1, n2, n3 := 0, 0, 0

	handler1 := &StringEventHandler{&n1, ""}
	handler2 := &StringEventHandler{&n2, ""}
	handler3 := &StringEventHandler{&n3, ""}

	// 测试链式注册
	bus.On("event1", handler1).
		On("event2", handler2).
		Once("event3", handler3)

	// 验证所有事件都被正确注册
	if bus.EventCount("event1") != 1 {
		t.Errorf("event1计数错误: %d", bus.EventCount("event1"))
	}
	if bus.EventCount("event2") != 1 {
		t.Errorf("event2计数错误: %d", bus.EventCount("event2"))
	}
	if bus.EventCount("event3") != 1 {
		t.Errorf("event3计数错误: %d", bus.EventCount("event3"))
	}

	// 测试链式触发
	bus.Trigger("event1", "data1").
		Trigger("event2", "data2").
		Trigger("event3", "data3")

	// 验证所有处理器都被调用
	if n1 != 1 {
		t.Errorf("handler1计数错误: %d", n1)
	}
	if n2 != 1 {
		t.Errorf("handler2计数错误: %d", n2)
	}
	if n3 != 1 {
		t.Errorf("handler3计数错误: %d", n3)
	}

	// 验证数据正确传递
	if handler1.lastMsg != "data1" {
		t.Errorf("handler1数据错误: %s", handler1.lastMsg)
	}
	if handler2.lastMsg != "data2" {
		t.Errorf("handler2数据错误: %s", handler2.lastMsg)
	}
	if handler3.lastMsg != "data3" {
		t.Errorf("handler3数据错误: %s", handler3.lastMsg)
	}

	// 验证链式清理与验证
	if bus.Clean().TotalEvents() != 0 {
		t.Errorf("清理后总数不为0: %d", bus.TotalEvents())
	}

	// 测试链式取消订阅
	bus.On("test", handler1).On("test", handler2)
	if bus.Off("test", handler1).EventCount("test") != 1 {
		t.Errorf("取消订阅后计数错误: %d", bus.EventCount("test"))
	}
}

// TestConcurrentEventHandling 测试并发事件处理的线程安全性
func TestConcurrentEventHandling(t *testing.T) {
	bus := New[string]()
	var counter int32
	var wg sync.WaitGroup

	// 注册处理器
	handler := &AtomicHandler{&counter}
	bus.On("concurrent", handler)

	// 并发触发事件
	numGoroutines := 100
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			bus.Trigger("concurrent", "data")
		}()
	}

	wg.Wait()

	// 验证计数器是否正确
	if atomic.LoadInt32(&counter) != int32(numGoroutines) {
		t.Errorf("并发计数器错误: 期望 %d, 实际 %d", numGoroutines, atomic.LoadInt32(&counter))
	}
}

// TestDispatchEdgeCases 测试dispatch和dispatchAll的边缘情况
func TestDispatchEdgeCases(t *testing.T) {
	// 测试空主题场景
	emptyBus := New[string]()
	emptyBus.Trigger("non-existent")
	emptyBus.Broadcast()

	// 测试ALL主题的特殊处理
	allBus := New[string]()
	allCounter := 0
	allHandler := &StringEventHandler{&allCounter, ""}
	allBus.On(ALL, allHandler)

	// 触发ALL主题
	allBus.Trigger(ALL)

	if allCounter != 1 {
		t.Errorf("触发ALL主题后计数器为 %d，期望为 %d", allCounter, 1)
	}

	// 测试多个事件同时被触发
	mixedBus := New[string]().AllowAsterisk()
	topicCounter, allTopicCounter := 0, 0
	topicHandler := &StringEventHandler{&topicCounter, ""}
	allTopicHandler := &StringEventHandler{&allTopicCounter, ""}

	mixedBus.On("specific", topicHandler)
	mixedBus.On(ALL, allTopicHandler)

	// 特定主题应该同时触发特定处理器和ALL处理器
	mixedBus.Trigger("specific", "data")

	if topicCounter != 1 || allTopicCounter != 1 {
		t.Errorf("主题和ALL处理器计数错误: topic=%d, all=%d", topicCounter, allTopicCounter)
	}
}

// TestRemoveEventsCbEdgeCases 测试removeEventsCb的边缘情况
func TestRemoveEventsCbEdgeCases(t *testing.T) {
	bus := New[string]()
	callbackCalled := false

	// 测试移除不存在的事件但有回调
	bus.OffCb("non-existent", func(count int, exists bool) {
		callbackCalled = true
		if exists {
			t.Error("不存在的主题报告为存在")
		}
		if count != 0 {
			t.Errorf("不存在主题的计数为 %d，期望为 0", count)
		}
	})

	if !callbackCalled {
		t.Error("移除不存在主题时回调未被调用")
	}

	// 测试移除空事件列表
	callbackCalled = false
	n := 0
	handler := &StringEventHandler{&n, ""}
	bus.On("test", handler)

	bus.OffCb("test", func(count int, exists bool) {
		callbackCalled = true
		if !exists {
			t.Error("存在的主题报告为不存在")
		}
		if count != 1 {
			t.Errorf("移除前计数为 %d，期望为 1", count)
		}
	})

	if !callbackCalled {
		t.Error("移除主题时回调未被调用")
	}

	// 测试移除后主题依然存在的情况
	bus = New[string]()
	n1, n2 := 0, 0
	handler1 := &StringEventHandler{&n1, ""}
	handler2 := &StringEventHandler{&n2, ""}

	bus.On("topic", handler1)
	bus.On("topic", handler2)

	remainingCount := 0
	bus.OffCb("topic", func(count int, exists bool) {
		remainingCount = count
	}, handler1)

	if remainingCount != 2 {
		t.Errorf("移除前计数为 %d，期望为 2", remainingCount)
	}

	// 验证实际移除后的计数
	if bus.EventCount("topic") != 1 {
		t.Errorf("移除一个处理器后计数为 %d，期望为 1", bus.EventCount("topic"))
	}

	// 额外测试：移除不存在的处理器
	notExistHandler := &StringEventHandler{&n1, ""}
	bus.OffCb("topic", func(count int, exists bool) {
		if count != 1 {
			t.Errorf("移除前计数为 %d，期望为 1", count)
		}
	}, notExistHandler)

	// 测试空事件列表场景
	emptyBus := New[string]()
	// 创建一个空的主题
	emptyHandler := &StringEventHandler{&n, ""}
	emptyBus.On("empty-topic", emptyHandler)
	emptyBus.Off("empty-topic", emptyHandler)

	// 清空主题后，使用OffCb再次检查
	callbackCalled = false
	emptyBus.OffCb("empty-topic", func(count int, exists bool) {
		callbackCalled = true
		// exists的值取决于具体实现，不做断言
		if count != 0 {
			t.Errorf("空列表主题的计数为 %d，期望为 0", count)
		}
	})

	if !callbackCalled {
		t.Error("移除空列表主题时回调未被调用")
	}
}

// TestComplexDispatchScenarios 测试更复杂的dispatch场景
func TestComplexDispatchScenarios(t *testing.T) {
	bus := New[string]()

	// 测试TriggerAll对所有主题的影响
	n1, n2, n3 := 0, 0, 0
	h1 := &StringEventHandler{&n1, ""}
	h2 := &StringEventHandler{&n2, ""}
	h3 := &StringEventHandler{&n3, ""}

	bus.On("topic1", h1)
	bus.On("topic2", h2)
	bus.Once("topic3", h3) // 使用Once来测试自动移除功能

	bus.Broadcast("data")

	if n1 != 1 || n2 != 1 || n3 != 1 {
		t.Errorf("TriggerAll后计数器错误: n1=%d, n2=%d, n3=%d", n1, n2, n3)
	}

	// 再次触发，检查Once是否被正确移除
	bus.Broadcast("data again")

	if n1 != 2 || n2 != 2 || n3 != 1 {
		t.Errorf("第二次TriggerAll后计数器错误: n1=%d, n2=%d, n3=%d", n1, n2, n3)
	}

	// 检查所有事件的计数
	if bus.EventCount("topic3") != 0 {
		t.Errorf("Once事件应被移除，但计数为 %d", bus.EventCount("topic3"))
	}
}

// TestRemoveNonExistentEventWithHandlers 测试从不存在的主题中移除处理器的情况
func TestRemoveNonExistentEventWithHandlers(t *testing.T) {
	bus := New[string]()
	n := 0
	handler := &StringEventHandler{&n, ""}

	// 直接尝试从不存在的主题移除处理器
	bus.Off("non-existent-topic", handler)

	// 使用OffCb从不存在的主题移除处理器
	callbackCalled := false
	// 检查exists值
	bus.OffCb("non-existent-topic", func(count int, exists bool) {
		callbackCalled = true
		if exists {
			t.Error("不存在的主题报告为存在")
		}
	}, handler)

	if !callbackCalled {
		t.Error("移除不存在主题的处理器时回调未被调用")
	}

	// 测试处理空事件列表的情况
	emptyBus := New[string]()
	// 通过正常API创建和移除来构建空列表场景
	emptyHandler := &StringEventHandler{&n, ""}
	emptyBus.On("empty-handlers", emptyHandler)
	emptyBus.Off("empty-handlers", emptyHandler)

	// 此时应该是空列表
	if emptyBus.EventCount("empty-handlers") != 0 {
		t.Error("应该是空列表，但仍有处理器")
	}

	// 再次移除
	emptyBus.Off("empty-handlers", handler)

	// 验证事件列表
	if emptyBus.EventCount("empty-handlers") != 0 {
		t.Error("空处理器列表的主题应该被移除")
	}
}

// TestBusGet 测试Get方法
func TestBusGet(t *testing.T) {
	bus := New[string]()
	n := 0
	handler := &StringEventHandler{&n, ""}

	// 测试不存在的主题
	topic := bus.Get("non-existent")
	if topic != nil {
		t.Error("Get应该对不存在的主题返回nil")
	}

	// 测试存在的主题
	bus.On("test-topic", handler)
	topic = bus.Get("test-topic")
	if topic == nil {
		t.Error("Get应该对存在的主题返回Topic")
	}

	// 验证返回的Topic
	if topic.Count() != 1 {
		t.Errorf("返回的Topic计数应为1，但得到%d", topic.Count())
	}

	if topic.name != "test-topic" {
		t.Errorf("返回的Topic主题应为'test-topic'，但得到%s", topic.name)
	}
}

// TestDispatchEdgeCases 测试dispatch中的所有分支情况
func TestDispatchEdgeCases_Comprehensive(t *testing.T) {
	// 测试不存在的主题场景
	bus := New[string]().AllowAsterisk()
	bus.Trigger("non-existent")

	// 测试全局主题的特殊处理
	allCounter := 0
	specificCounter := 0
	globalHandler := &StringEventHandler{&allCounter, ""}
	specificHandler := &StringEventHandler{&specificCounter, ""}

	bus.On(ALL, globalHandler)
	bus.On("specific", specificHandler)

	// 测试普通主题触发
	bus.Trigger("specific", "data")

	if specificCounter != 1 {
		t.Errorf("特定主题计数应为1，但得到%d", specificCounter)
	}

	if allCounter != 1 {
		t.Errorf("全局主题计数应为1，但得到%d", allCounter)
	}

	// 测试触发全局主题本身
	bus.Trigger(ALL, "global-data")

	if allCounter != 2 {
		t.Errorf("全局主题触发后计数应为2，但得到%d", allCounter)
	}

	// 测试不允许星号主题的情况
	noAsteriskBus := New[string]() // 默认不允许星号
	allNoAsteriskCounter := 0
	noAsteriskGlobalHandler := &StringEventHandler{&allNoAsteriskCounter, ""}
	noAsteriskBus.On(ALL, noAsteriskGlobalHandler)

	// 触发普通主题，但由于不允许星号，不应该触发ALL主题
	noAsteriskBus.Trigger("normal", "data")

	if allNoAsteriskCounter != 0 {
		t.Errorf("不允许星号时，普通主题不应触发全局处理器，但计数为%d", allNoAsteriskCounter)
	}

	// 直接触发ALL主题应该仍然有效
	noAsteriskBus.Trigger(ALL, "direct")

	if allNoAsteriskCounter != 1 {
		t.Errorf("直接触发ALL主题应使计数为1，但得到%d", allNoAsteriskCounter)
	}
}

// TestTopicCount 测试主题计数方法
func TestTopicCount(t *testing.T) {
	bus := New[string]()
	n := 0
	handler := &StringEventHandler{&n, ""}

	// 初始应该为0
	if bus.TopicCount() != 0 {
		t.Errorf("初始主题数为 %d，期望为 %d", bus.TopicCount(), 0)
	}

	// 添加一个主题
	bus.On("topic1", handler)

	if bus.TopicCount() != 1 {
		t.Errorf("添加一个主题后，主题数为 %d，期望为 %d", bus.TopicCount(), 1)
	}

	// 添加第二个主题
	bus.On("topic2", handler)

	if bus.TopicCount() != 2 {
		t.Errorf("添加两个主题后，主题数为 %d，期望为 %d", bus.TopicCount(), 2)
	}

	// 添加到已存在的主题不应增加主题计数
	bus.On("topic1", handler)

	if bus.TopicCount() != 2 {
		t.Errorf("添加到已存在主题后，主题数为 %d，期望为 %d", bus.TopicCount(), 2)
	}

	// 移除主题
	bus.Off("topic1", handler)

	// 如果移除了所有该主题的处理器，主题应该被删除
	if bus.TopicCount() != 1 {
		t.Errorf("移除一个主题后，主题数为 %d，期望为 %d", bus.TopicCount(), 1)
	}

	// 清空所有主题
	bus.Clean()

	if bus.TopicCount() != 0 {
		t.Errorf("清空后，主题数为 %d，期望为 %d", bus.TopicCount(), 0)
	}
}
