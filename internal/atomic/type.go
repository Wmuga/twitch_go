package atomic

import "sync"

type Type[T any] struct {
	mux  *sync.Mutex
	item T
}

func NewType[T any]() *Type[T] {
	return &Type[T]{
		mux: &sync.Mutex{},
	}
}

func (t *Type[T]) Load() T {
	t.mux.Lock()
	defer t.mux.Unlock()
	return t.item
}

func (t *Type[T]) Store(item T) {
	t.mux.Lock()
	defer t.mux.Unlock()
	t.item = item
}
