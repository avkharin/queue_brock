package main

import (
	"sync"
	"time"
)

type Queue struct {
	messages []string
	waiters  []chan string
	mu       sync.Mutex
}

func (q *Queue) Enqueue(msg string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.waiters) > 0 {
		waiter := q.waiters[0]
		q.waiters = q.waiters[1:]
		waiter <- msg
		close(waiter)
	} else {
		q.messages = append(q.messages, msg)
	}
}

func (q *Queue) Dequeue(timeout time.Duration) (string, bool) {

}
