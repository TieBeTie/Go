//go:build !solution

package keylock

import (
	"sort"
	"sync"
)

type KeyLock struct {
	mu      sync.Mutex
	locks   map[string]bool
	waiters map[string][]chan struct{}
}

func New() *KeyLock {
	return &KeyLock{
		locks:   make(map[string]bool),
		waiters: make(map[string][]chan struct{}),
	}
}

func safeClose(ch chan struct{}) {
	defer func() {
		recover()
	}()
	close(ch)
}

func (l *KeyLock) LockKeys(keys []string, cancel <-chan struct{}) (canceled bool, unlock func()) {
	ks := make([]string, len(keys))
	copy(ks, keys)
	sort.Strings(ks)

	for {
		waitCh := make(chan struct{})
		var conflictKeys []string

		l.mu.Lock()
		conflict := false
		for _, k := range ks {
			if l.locks[k] {
				conflict = true
				conflictKeys = append(conflictKeys, k)
			}
		}

		if !conflict {
			for _, k := range ks {
				l.locks[k] = true
			}
			l.mu.Unlock()

			unlocked := false
			return false, func() {
				l.mu.Lock()
				defer l.mu.Unlock()

				if unlocked {
					return
				}
				unlocked = true

				for _, k := range ks {
					l.locks[k] = false

					if waiters := l.waiters[k]; len(waiters) > 0 {
						safeClose(waiters[0])
						l.waiters[k] = waiters[1:]
						if len(l.waiters[k]) == 0 {
							delete(l.waiters, k)
						}
					}
				}
			}
		}

		for _, k := range conflictKeys {
			if _, ok := l.waiters[k]; !ok {
				l.waiters[k] = []chan struct{}{}
			}
			l.waiters[k] = append(l.waiters[k], waitCh)
		}
		l.mu.Unlock()

		select {
		case <-waitCh:
			continue
		case <-cancel:
			l.mu.Lock()
			for _, k := range conflictKeys {
				waiters := l.waiters[k]
				for i, ch := range waiters {
					if ch == waitCh {
						l.waiters[k] = append(waiters[:i], waiters[i+1:]...)
						if len(l.waiters[k]) == 0 {
							delete(l.waiters, k)
						}
						break
					}
				}
			}
			l.mu.Unlock()

			safeClose(waitCh)
			return true, func() {}
		}
	}
}

func isClosed(ch <-chan struct{}) bool {
	select {
	case <-ch:
		return true
	default:
		return false
	}
}
