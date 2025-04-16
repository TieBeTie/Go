//go:build !solution

package pubsub

import (
	"context"
	"sync"
)

type MySubscription struct {
	done            chan struct{}
	unsubscribeFunc func()
}

func (s *MySubscription) Unsubscribe() {
	close(s.done)
	if s.unsubscribeFunc != nil {
		s.unsubscribeFunc()
	}
}

type MyPubSub struct {
	mu     sync.RWMutex
	subs   map[string]map[chan<- interface{}]struct{}
	closed bool
	wg     sync.WaitGroup
}

func NewPubSub() PubSub {
	return &MyPubSub{
		subs: make(map[string]map[chan<- interface{}]struct{}),
	}
}

func (p *MyPubSub) Subscribe(subj string, cb MsgHandler) (Subscription, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil, context.Canceled
	}

	msgCh := make(chan interface{}, 100)
	doneCh := make(chan struct{})

	if _, ok := p.subs[subj]; !ok {
		p.subs[subj] = make(map[chan<- interface{}]struct{})
	}

	p.subs[subj][msgCh] = struct{}{}
	p.wg.Add(1)

	go func() {
		defer p.wg.Done()

		for {
			select {
			case msg, ok := <-msgCh:
				if !ok {
					return
				}
				cb(msg)
			case <-doneCh:
				return
			}
		}
	}()

	subscription := &MySubscription{
		done: doneCh,
		unsubscribeFunc: func() {
			p.mu.Lock()
			defer p.mu.Unlock()

			if topicSubs, ok := p.subs[subj]; ok {
				delete(topicSubs, msgCh)
				if len(topicSubs) == 0 {
					delete(p.subs, subj)
				}
			}

			close(msgCh)
		},
	}

	return subscription, nil
}

func (p *MyPubSub) Publish(subj string, msg interface{}) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		return context.Canceled
	}

	subs, ok := p.subs[subj]
	if !ok || len(subs) == 0 {
		return nil
	}

	for ch := range subs {
		ch <- msg
	}

	return nil
}

func (p *MyPubSub) Close(ctx context.Context) error {
	var channels []chan<- interface{}

	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil
	}

	p.closed = true

	for _, topicSubs := range p.subs {
		for ch := range topicSubs {
			channels = append(channels, ch)
		}
	}

	p.subs = make(map[string]map[chan<- interface{}]struct{})
	p.mu.Unlock()

	for _, ch := range channels {
		close(ch)
	}

	done := make(chan struct{})

	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return nil
	}
}
