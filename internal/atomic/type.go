package atomic

import "sync"

// Atomic generic type.
// Use NewType to create new instance.
type Type[T any] struct {
	mux  *sync.Mutex
	item T
}

// Creates new instance of genetic atomic type
func NewType[T any]() *Type[T] {
	return &Type[T]{
		mux: &sync.Mutex{},
	}
}

// Get item value
func (t *Type[T]) Load() T {
	t.mux.Lock()
	defer t.mux.Unlock()
	return t.item
}

// Load new item value
func (t *Type[T]) Store(item T) {
	t.mux.Lock()
	defer t.mux.Unlock()
	t.item = item
}
