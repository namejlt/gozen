package event

import (
	"reflect"
	"sync"
	"time"

	"github.com/namejlt/gozen/concurrent"
	"github.com/namejlt/gozen/storage"
)

var (
	RollBus = NewLocalBus(20)
)

type publicReq struct {
	Topic string
	Args  []reflect.Value
}

// LocalEventBus is the default Bus implementation — an in-process pub/sub
// with async dispatch and concurrency control.
type LocalEventBus struct {
	sync.Mutex
	topics map[string][]reflect.Value
	done   chan struct{}
	once   sync.Once
	limit  concurrent.Concurrenter
	queue  *storage.Link
}

// NewLocalBus creates a LocalEventBus with the given max concurrent goroutines.
func NewLocalBus(max int) *LocalEventBus {
	e := &LocalEventBus{
		topics: make(map[string][]reflect.Value),
		done:   make(chan struct{}),
		limit:  concurrent.NewSimpleConcurrentLimit(max),
		queue:  new(storage.Link),
	}
	go e.Work()
	return e
}

// Stop implements Bus.Stop.
func (e *LocalEventBus) Stop() {
	e.once.Do(func() {
		close(e.done)
	})
}

// Subscribe implements Bus.Subscribe.
func (e *LocalEventBus) Subscribe(topic string, fn any) {
	if topic != "" && reflect.TypeOf(fn).Kind() == reflect.Func {
		callback := reflect.ValueOf(fn)
		e.Lock()
		e.topics[topic] = append(e.topics[topic], callback)
		e.Unlock()
	}
}

// Unsubscribe implements Bus.Unsubscribe.
func (e *LocalEventBus) Unsubscribe(topic string, fn any) {
	if topic != "" && reflect.TypeOf(fn).Kind() == reflect.Func {
		callback := reflect.ValueOf(fn)
		e.Lock()
		defer e.Unlock()
		var channels []reflect.Value
		for _, c := range e.topics[topic] {
			if c != callback {
				channels = append(channels, c)
			}
		}
		e.topics[topic] = channels
	}
}

// Publish implements Bus.Publish.
func (e *LocalEventBus) Publish(topic string, args ...any) {
	if topic != "" {
		passedArgs := make([]reflect.Value, 0)
		for _, arg := range args {
			passedArgs = append(passedArgs, reflect.ValueOf(arg))
		}
		e.queue.Push(&publicReq{
			Topic: topic,
			Args:  passedArgs,
		})
	}
}

// Work is the main event dispatch loop.
func (e *LocalEventBus) Work() {
	delays := []int{10, 25, 50, 75, 100, 150, 250}
	curIdx := 0
	maxIdx := len(delays) - 1

	for {
		select {
		case <-e.done:
			return
		default:
			element, found := e.queue.Pop()
			if !found {
				if curIdx < maxIdx {
					curIdx++
				}
				time.Sleep(time.Duration(delays[curIdx]) * time.Millisecond)
				continue
			}
			curIdx = 0
			p := element.(*publicReq)
			e.Lock()
			callbacks, found := e.topics[p.Topic]
			e.Unlock()
			if !found {
				continue
			}
			for _, fn := range callbacks {
				e.limit.Acquire()
				go e.execute(fn, p.Args)
			}
		}
	}
}

func (e *LocalEventBus) execute(callback reflect.Value, args []reflect.Value) {
	defer e.limit.Release()
	callback.Call(args)
}
