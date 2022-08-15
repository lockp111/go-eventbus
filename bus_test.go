package eventbus

import (
	"log"
	"sync"
	"testing"
)

type N struct {
	i *int
	s string
}

func (n *N) Dispatch(data ...string) {
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

func (fn *FN) Dispatch(data ...input) {
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
	o.Trigger("foo")

	o.Off("foo", onFoo1).Off("foo", onFoo2).On("foo", onFoo1)
	o.Trigger("foo")

	if n != 3 {
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
