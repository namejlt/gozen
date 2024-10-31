package gozen

import (
	"reflect"
	"sync"
	"time"

	"github.com/namejlt/gozen/concurrent"
	"github.com/namejlt/gozen/storage"
)

var (
	RollBus = NewEventBus(20)
)

type publicReq struct {
	Topic string
	Args  []reflect.Value
}

type EventBus struct {
	sync.Mutex
	topics map[string][]reflect.Value
	done   chan struct{} //停止信号
	once   sync.Once
	limit  concurrent.Concurrenter //并发数限制
	queue  *storage.Link           //事件存储结构（控制事件调度顺序,队列、链表、栈）
}

func NewEventBus(max int) *EventBus {
	e := &EventBus{
		topics: make(map[string][]reflect.Value),
		done:   make(chan struct{}),
		limit:  concurrent.NewSimpleConcurrentLimit(max),
		queue:  new(storage.Link),
	}
	go e.Work()
	return e
}

// 关闭事件中心
func (e *EventBus) Stop() {
	e.once.Do(func() {
		close(e.done)
	})
}

// 订阅
func (e *EventBus) Subscribe(topic string, fn interface{}) {
	if topic != "" && reflect.TypeOf(fn).Kind() == reflect.Func {
		callback := reflect.ValueOf(fn)
		e.Lock()
		e.topics[topic] = append(e.topics[topic], callback)
		e.Unlock()

	}
}

// 取消订阅
func (e *EventBus) Unsubscribe(topic string, fn interface{}) {
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

// 发布
func (e *EventBus) Publish(topic string, args ...interface{}) {
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

// 事件调度中心
func (e *EventBus) Work() {
	curIdx := 0
	maxIdx := 6
	delays := []int{10, 25, 50, 75, 100, 150, 250}

	for {
		select {
		case <-e.done:
			return
		default:
			element, found := e.queue.Pop()
			if !found {
				//层级等待
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

func (e *EventBus) execute(callback reflect.Value, args []reflect.Value) {
	defer e.limit.Release()
	callback.Call(args)
}
