package atomic

import (
	"errors"
	"sync"
)

type List[T any] struct {
	mux  *sync.Mutex
	list []T
}

var (
	ErrListEmpty = errors.New("empty list")
)

func NewList[T any]() *List[T] {
	return &List[T]{
		mux:  &sync.Mutex{},
		list: make([]T, 0),
	}
}

func (l *List[T]) Push(items ...T) {
	l.mux.Lock()
	defer l.mux.Unlock()
	l.list = append(l.list, items...)
}

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

func (l *List[T]) Count() int {
	l.mux.Lock()
	defer l.mux.Unlock()

	return len(l.list)
}

func (l *List[T]) Elements() []T {
	return l.list
}

func (l *List[T]) At(index int) T {
	l.mux.Lock()
	defer l.mux.Unlock()

	return l.list[index]
}
