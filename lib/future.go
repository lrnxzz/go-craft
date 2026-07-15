package lib

import (
	"context"
	"sync"
)

type Future[T any] struct {
	once  sync.Once
	done  chan struct{}
	value T
	err   error
}

func NewFuture[T any]() *Future[T] {
	return &Future[T]{
		done: make(chan struct{}),
	}
}

func FailedFuture[T any](err error) *Future[T] {
	future := NewFuture[T]()
	var zero T
	future.Complete(zero, err)

	return future
}

func (f *Future[T]) Complete(value T, err error) {
	f.once.Do(func() {
		f.value = value
		f.err = err
		close(f.done)
	})
}

func (f *Future[T]) Done() <-chan struct{} {
	return f.done
}

func (f *Future[T]) Wait(ctx context.Context) (T, error) {
	select {
	case <-f.done:
		return f.value, f.err
	case <-ctx.Done():
		var zero T

		return zero, ctx.Err()
	}
}
