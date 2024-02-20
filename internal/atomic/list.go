package atomic

import (
	"errors"
	"sync"
)

// List with atomic operations and generics.
// Use NewList to create new instance
type List[T any] struct {
	mux  *sync.Mutex
	list []T
}

var (
	ErrListEmpty = errors.New("empty list")
)

// Create new List
func NewList[T any]() *List[T] {
	return &List[T]{
		mux:  &sync.Mutex{},
		list: make([]T, 0),
	}
}

// Add new element to end of list
func (l *List[T]) Push(items ...T) {
	l.mux.Lock()
	defer l.mux.Unlock()
	l.list = append(l.list, items...)
}

// Remove and return last element of list. Returns error if empty
func (l *List[T]) Pop() (item T, err error) {
	l.mux.Lock()
	defer l.mux.Unlock()

	if len(l.list) == 0 {
		return item, ErrListEmpty
	}

	res := l.list[len(l.list)-1]
	l.list = l.list[:len(l.list)-1]
	return res, nil
}

// Remove and return first element of list. Returns error if empty
func (l *List[T]) Shift() (item T, err error) {
	l.mux.Lock()
	defer l.mux.Unlock()

	if len(l.list) == 0 {
		return item, ErrListEmpty
	}

	res := l.list[0]
	l.list = l.list[1:]
	return res, nil
}

// Get elements count
func (l *List[T]) Count() int {
	l.mux.Lock()
	defer l.mux.Unlock()

	return len(l.list)
}

// Get all elements as array
func (l *List[T]) Elements() []T {
	l.mux.Lock()
	defer l.mux.Unlock()
	return l.list
}

// Get element at {index}
func (l *List[T]) At(index int) T {
	l.mux.Lock()
	defer l.mux.Unlock()

	return l.list[index]
}
