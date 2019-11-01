# go-eventbus

An golang event bus model

## Usage

### go get
```go
go get -u github.com/lockp111/go-eventbus
```

#### New()

Create a new bus struct reference

```go
bus := eventbus.New()
```

### On(topic string, e ...Event)

Subscribe event

```go
type ready struct{
}

func (e ready) Dispatch(msg interface{}){
    fmt.Println("I am ready!")
}

bus.On("ready", &ready{})
```

You can also subscribe multiple events for example:

```go
type run struct{
}

func (e run) Dispatch(msg interface{}){
    fmt.Println("I am run!")
}

bus.On("ready", &ready{}, &ready{}).On("run", &run{})
```

### Off(topic string, e ...Event)

Unsubscribe event

```go
e := &ready{}
bus.On("ready", e)
bus.Off("ready", e)
```

You can also unsubscribe multiple events for example:

```go
e1 := &ready{}
e2 := &ready{}
bus.On("ready", e1, e2)
bus.Off("ready", e1, e2)
```

You can unsubscribe all events for example:

```go
bus.On("ready", &ready{}, &ready{})
bus.Off("ready")
```

You can unsubscribe all topics for example:

```go
bus.On("ready", &ready{}, &ready{})
bus.On("run", &run{})
bus.Off(ALL_TOPICS)
```

### Trigger(topic string, msg ...interface{})

Dispatch events

```go
bus.Trigger("ready")
```

You can also dispatch multiple events for example:

```go
bus.Trigger("ready", &struct{"1"}, &struct{"2"})
```

You can also dispatch all events for example:

```go
bus.Trigger(ALL_TOPICS, &struct{"1"})
```