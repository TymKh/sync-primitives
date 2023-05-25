package patterns

import (
	"context"
	"sync"
	"sync/atomic"

	"golang.org/x/sync/semaphore"
)

type TimedMutex struct {
	ch chan struct{}
}

func NewTimedMutex() *TimedMutex {
	return &TimedMutex{ch: make(chan struct{}, 1)}
}

func (tm *TimedMutex) Lock() {
	tm.ch <- struct{}{}
}

func (tm *TimedMutex) Unlock() {
	<-tm.ch
}

func (tm *TimedMutex) TryLock() bool {
	select {
	case tm.ch <- struct{}{}:
		return true
	default:
		return false
	}
}

func (tm *TimedMutex) TryLockWithContext(ctx context.Context) bool {
	select {
	case tm.ch <- struct{}{}:
		return true
	case <-ctx.Done():
		return false
	}
}

const (
	stateWriteLock = iota - 1
	stateNoLock
)

type TimedRWMutex struct {
	stateAccessSynchronizer *semaphore.Weighted
	state                   atomic.Int32
	stateChangeMu           sync.RWMutex
	stateChangeCh           chan struct{}
}

func NewTimedRWMutex() *TimedRWMutex {
	return &TimedRWMutex{
		state:                   atomic.Int32{},
		stateAccessSynchronizer: semaphore.NewWeighted(1),
		stateChangeCh:           make(chan struct{}),
	}
}

func (m *TimedRWMutex) getChangeChan() <-chan struct{} {
	m.stateChangeMu.RLock()
	defer m.stateChangeMu.RUnlock()

	return m.stateChangeCh
}

func (m *TimedRWMutex) broadcastStateChange() {
	newCh := make(chan struct{})

	m.stateChangeMu.Lock()
	ch := m.stateChangeCh
	m.stateChangeCh = newCh
	m.stateChangeMu.Unlock()

	close(ch)
}

func (m *TimedRWMutex) tryLock(ctx context.Context) bool {
	for {
		changeCh := m.getChangeChan()

		if m.state.CompareAndSwap(stateNoLock, stateWriteLock) {
			return true
		}

		if ctx == nil {
			return false
		}

		select {
		case <-ctx.Done():
			return false
		case <-changeCh:
		}
	}
}

// TryLockWithContext attempts to acquire the lock, blocking until resources
// are available or ctx is done (timeout or cancellation).
func (m *TimedRWMutex) TryLockWithContext(ctx context.Context) bool {
	if err := m.stateAccessSynchronizer.Acquire(ctx, 1); err != nil {
		return false
	}

	defer m.stateAccessSynchronizer.Release(1)

	return m.tryLock(ctx)
}

func (m *TimedRWMutex) Lock() {
	ctx := context.Background()

	m.TryLockWithContext(ctx)
}

func (m *TimedRWMutex) TryLock() bool {
	if !m.stateAccessSynchronizer.TryAcquire(1) {
		return false
	}

	defer m.stateAccessSynchronizer.Release(1)

	return m.tryLock(nil)
}

// Unlock releases the lock.
func (m *TimedRWMutex) Unlock() {
	ok := m.state.CompareAndSwap(stateWriteLock, stateNoLock)
	if !ok {
		panic("Unlock not locked mutex")
	}

	m.broadcastStateChange()
}

func (m *TimedRWMutex) rTryLock(ctx context.Context) bool {
	for {
		broker := m.getChangeChan()
		n := m.state.Load()
		switch {
		case n >= 0:
			if m.state.CompareAndSwap(n, n+1) {
				return true
			}
		}

		if ctx == nil {
			return false
		}

		select {
		case <-ctx.Done():
			return false
		default:
			if n >= 0 {
				continue
			}
		}

		select {
		case <-ctx.Done():
			return false
		case <-broker:
		}
	}
}

// RTryLockWithContext attempts to acquire the read lock, blocking until resources
// are available or ctx is done (timeout or cancellation).
func (m *TimedRWMutex) RTryLockWithContext(ctx context.Context) bool {
	if err := m.stateAccessSynchronizer.Acquire(ctx, 1); err != nil {
		// Acquire failed due to timeout or cancellation
		return false
	}

	m.stateAccessSynchronizer.Release(1)

	return m.rTryLock(ctx)
}

// RLock acquires the read lock.
// If it is currently held by others writing, RLock will wait until it has a chance to acquire it.
func (m *TimedRWMutex) RLock() {
	ctx := context.Background()

	m.RTryLockWithContext(ctx)
}

// RTryLock attempts to acquire the read lock without blocking.
// Return false if someone is writing it now.
func (m *TimedRWMutex) RTryLock() bool {
	if !m.stateAccessSynchronizer.TryAcquire(1) {
		return false
	}

	m.stateAccessSynchronizer.Release(1)

	return m.rTryLock(nil)
}

// RUnlock releases the read lock.
func (m *TimedRWMutex) RUnlock() {
	n := m.state.Add(-1)
	if n == stateWriteLock {
		panic("RUnlock failed")
	}
	if n >= 0 {
		m.broadcastStateChange()
	}
}
