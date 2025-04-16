//go:build !solution

package dupcall

import (
	"context"
	"sync"
)

type Call struct {
	mu      sync.Mutex
	result  interface{}
	err     error
	ready   chan struct{}
	waiting int
	cancel  context.CancelFunc
	running bool
}

func (c *Call) Do(ctx context.Context, cb func(context.Context) (interface{}, error)) (interface{}, error) {
	c.mu.Lock()
	if c.running {
		c.waiting++
		ready := c.ready
		c.mu.Unlock()

		select {
		case <-ready:
			return c.result, c.err
		case <-ctx.Done():
			c.mu.Lock()
			c.waiting--
			if c.waiting == 0 && c.cancel != nil {
				c.cancel()
			}
			c.mu.Unlock()
			return nil, ctx.Err()
		}
	}

	// Первый, кто зашёл
	c.running = true
	c.waiting = 1
	c.ready = make(chan struct{})
	cbCtx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel
	c.mu.Unlock()

	go func() {
		result, err := cb(cbCtx)
		c.mu.Lock()
		c.result = result
		c.err = err
		close(c.ready)
		c.running = false
		c.cancel = nil
		c.mu.Unlock()
	}()

	select {
	case <-c.ready:
		c.mu.Lock()
		c.waiting--
		c.mu.Unlock()
		return c.result, c.err
	case <-ctx.Done():
		c.mu.Lock()
		c.waiting--
		if c.waiting == 0 && c.cancel != nil {
			c.cancel()
		}
		c.mu.Unlock()
		return nil, ctx.Err()
	}
}
