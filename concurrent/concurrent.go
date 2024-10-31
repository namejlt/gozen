package concurrent

type Concurrenter interface {
	Acquire()
	Release()
}
