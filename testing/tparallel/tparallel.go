package tparallel

import (
	"sync"
	"time"
)

type T struct {
	mu           sync.Mutex
	isParallel   bool
	parent       *T
	wg           sync.WaitGroup
	barrier      chan struct{} // barrier for top-level parallel tests
	parentReady  chan struct{} // signals parent test function has completed
	childrenDone chan struct{} // signals all child tests have completed
}

func (t *T) Parallel() {
	t.mu.Lock()
	if t.isParallel {
		t.mu.Unlock()
		return
	}
	t.isParallel = true
	t.mu.Unlock()

	if t.parent != nil {
		// Wait for parent test function to complete
		<-t.parent.parentReady
	} else {
		// Wait for all sequential tests to complete
		<-t.barrier
	}
}

func (t *T) Run(subtest func(t *T)) {
	child := &T{
		parent:       t,
		barrier:      t.barrier,
		parentReady:  make(chan struct{}),
		childrenDone: make(chan struct{}),
	}

	t.wg.Add(1)

	// Run the subtest
	go func() {
		defer t.wg.Done()

		// Execute the subtest
		subtest(child)

		// Signal parent that this test function has completed
		close(child.parentReady)

		// Wait for all subtests to complete
		child.wg.Wait()

		// Signal that all children are done
		close(child.childrenDone)
	}()

	// If test doesn't go parallel, wait for it and all its subtests to complete
	for {
		child.mu.Lock()
		isParallel := child.isParallel
		child.mu.Unlock()

		if isParallel {
			break
		}

		// Check if the test completed
		select {
		case <-child.childrenDone:
			return
		case <-time.After(time.Millisecond):
			// Try again
			continue
		}
	}
}

func Run(topTests []func(t *T)) {
	barrier := make(chan struct{})
	var wg sync.WaitGroup

	// Start all top-level tests
	for _, tf := range topTests {
		test := &T{
			barrier:      barrier,
			parentReady:  make(chan struct{}),
			childrenDone: make(chan struct{}),
		}

		wg.Add(1)

		go func(tf func(*T), t *T) {
			defer wg.Done()

			// Run the test
			tf(t)

			// Signal test function has completed
			close(t.parentReady)

			// Wait for all subtests to complete
			t.wg.Wait()

			// Signal all children are done
			close(t.childrenDone)
		}(tf, test)

		// Wait for test to complete or go parallel
		for {
			test.mu.Lock()
			isParallel := test.isParallel
			test.mu.Unlock()

			if isParallel {
				// If test went parallel, proceed to next test
				break
			}

			// Check if the test completed
			select {
			case <-test.childrenDone:
				// Test completed without going parallel
				goto nextTest
			case <-time.After(time.Millisecond):
				// Try again
				continue
			}
		}

	nextTest:
		continue
	}

	// All sequential tests have completed, release parallel tests
	close(barrier)

	// Wait for all tests to complete
	wg.Wait()
}
