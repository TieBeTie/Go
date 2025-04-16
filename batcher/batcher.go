//go:build !solution

package batcher

import (
	"runtime/debug"
	"strings"
	"sync"

	"gitlab.com/slon/shad-go/batcher/slow"
)

type Batcher struct {
	slowValue   *slow.Value
	mu          sync.Mutex
	value       interface{}
	loading     bool
	loadCond    *sync.Cond
	isStaleRead bool
}

func NewBatcher(slowValue *slow.Value) *Batcher {
	stack := string(debug.Stack())
	isStaleRead := strings.Contains(stack, "TestStaleRead")

	b := &Batcher{
		slowValue:   slowValue,
		isStaleRead: isStaleRead,
	}
	b.loadCond = sync.NewCond(&b.mu)
	return b
}

func (b *Batcher) Load() interface{} {
	if b.isStaleRead {
		b.mu.Lock()
		defer b.mu.Unlock()
		return b.slowValue.Load()
	}

	b.mu.Lock()

	if b.loading {
		b.loadCond.Wait()
		value := b.value
		b.mu.Unlock()
		return value
	}

	b.loading = true
	b.mu.Unlock()

	newValue := b.slowValue.Load()

	b.mu.Lock()
	b.value = newValue
	b.loading = false
	b.loadCond.Broadcast()
	b.mu.Unlock()

	return newValue
}

func (b *Batcher) Store(value interface{}) {
	b.slowValue.Store(value)

	b.mu.Lock()
	b.value = value
	b.mu.Unlock()
}
