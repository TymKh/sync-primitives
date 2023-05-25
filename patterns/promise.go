package patterns

import (
	"context"
)

type FutureStatus string

//type Future[T any] struct {
//	ch        <-chan T
//	result    T
//	retrieved bool
//}
//
//func (f *Future[T]) Get() T {
//	f.Wait()
//	return f.result
//}
//
//func (f *Future[T]) Wait() {
//	a := <-f.ch
//	f.result = a
//	f.retrieved = true
//}
//
//func (f *Future[T]) WaitCtx(ctx context.Context) {
//
//}

type SharedFuture[T any] struct {
	ch     <-chan T
	done   chan struct{}
	result *T
}

func (sf *SharedFuture[T]) WaitCtx(ctx context.Context) bool {
	select {
	case r, closed := <-sf.ch:
		if closed {
			<-sf.done
			return true
		}
		sf.result = &r
		close(sf.done)
		return true
	case <-sf.done:
		return true
	case <-ctx.Done():
		return false
	}
}
func (sf *SharedFuture[T]) Wait() {
	sf.WaitCtx(context.Background())
}
func (sf *SharedFuture[T]) Get() T {
	sf.Wait()
	return *sf.result
}

func (sf *SharedFuture[T]) GetCtx(ctx context.Context) (*T, bool) {
	ok := sf.WaitCtx(ctx)
	if !ok {
		return nil, false
	}
	return sf.result, true
}

type Promise[T any] struct {
	ch chan T
}

func NewPromise[T any]() *Promise[T] {
	return &Promise[T]{ch: make(chan T, 1)}
}

func (p *Promise[T]) SetValue(v T) {
	p.ch <- v
	close(p.ch)
}
func (p *Promise[T]) GetSharedFuture() SharedFuture[T] {
	return SharedFuture[T]{ch: p.ch, done: make(chan struct{})}
}
