package main

import (
	"net/http"
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
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.messages) > 0 {
		msg := q.messages[0]
		q.messages = q.messages[1:]
		return msg, true
	}

	if timeout > 0 {
		waiter := make(chan string, 1)
		q.waiters = append(q.waiters, waiter)
		q.mu.Unlock()

		select {
		case msg := <-waiter:
			return msg, true
		case <-time.After(timeout):
			// Remove the chanal from list
			q.mu.Lock()
			for i, w := range q.waiters {
				if w == waiter {
					q.waiters = append(q.waiters[:i], q.waiters[i+1:]...)
					break
				}
			}
			q.mu.Unlock()
			return "", false
		}

	}

	return "", false
}

type Brocker struct {
	mp map[string]*Queue
	mu sync.Mutex
}

func NewBrocker() *Brocker {
	return &Brocker{
		mp: make(map[string]*Queue),
	}
}

func (this *Brocker) getQueue(name string) *Queue {
	this.mu.Lock()
	defer this.mu.Unlock()
	if _, exist := this.mp[name]; !exist {
		this.mp[name] = &Queue{}
	}
	return this.mp[name]
}

func handlePut(w http.ResponseWriter, r *http.Request) {
	queueName := r.URL.Path[1:]
	message := r.URL.Query().Get("v")
	if message == "" {
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	brocker.getQueue(queueName).Enqueue(message)
	w.WriteHeader(http.StatusOK)
}

var brocker = NewBrocker()
