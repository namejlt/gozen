package storage

import "sync"

type node struct {
	Next *node
	Data interface{}
}

type Link struct {
	sync.Mutex
	size       int
	head, tail *node
}

func (l *Link) Push(data interface{}) {
	element := new(node)
	l.Lock()
	defer l.Unlock()
	l.size++
	if l.head == nil {
		l.head = element
	}
	end := l.tail

	if end != nil {
		end.Next = element
	}
	l.tail = element
	element.Data = data
}

func (l *Link) Pop() (interface{}, bool) {
	l.Lock()
	defer l.Unlock()

	if l.head == nil {
		return nil, false
	}

	l.size--
	element := l.head
	l.head = element.Next
	if l.tail == element {
		l.tail = nil
	}
	element.Next = nil
	return element.Data, true
}

func (l *Link) Size() int {
	l.Lock()
	defer l.Unlock()
	return l.size
}
