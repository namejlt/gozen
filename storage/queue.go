package storage

type Storager interface {
	Push(interface{})
	Pop() (interface{}, bool)
}
