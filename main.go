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

}

func (q *Queue) Dequeue(timeout time.Duration) (string, bool) {

}
