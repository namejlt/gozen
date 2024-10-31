package concurrent

type SimpleConcurrent struct {
	c chan struct{}
}

func NewSimpleConcurrentLimit(max int) *SimpleConcurrent {
	s := &SimpleConcurrent{
		c: make(chan struct{}, max),
	}
	for i := 0; i < max; i++ {
		s.c <- struct{}{}
	}
	return s
}

// 获取运行权限
func (s *SimpleConcurrent) Acquire() {
	select {
	case <-s.c:
	}
}

// 是否运行权限
func (s *SimpleConcurrent) Release() {
	s.c <- struct{}{}
}
