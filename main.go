package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
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
	//	q.mu.Lock()
	//	defer q.mu.Unlock()

	if len(q.messages) > 0 {
		q.mu.Lock()
		msg := q.messages[0]
		q.messages = q.messages[1:]
		q.mu.Unlock()

		return msg, true
	}

	if timeout > 0 {
		waiter := make(chan string, 1)
		q.mu.Lock()
		q.waiters = append(q.waiters, waiter)
		q.mu.Unlock()

		select {
		case msg := <-waiter:
			return msg, true
		case <-time.After(timeout):
			// Remove the chanel from list
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

type Broker struct {
	mp map[string]*Queue
	mu sync.Mutex
}

func NewBroker() *Broker {
	return &Broker{
		mp: make(map[string]*Queue),
	}
}

func (broker *Broker) getQueue(name string) *Queue {
	broker.mu.Lock()
	defer broker.mu.Unlock()
	if _, exist := broker.mp[name]; !exist {
		broker.mp[name] = &Queue{}
	}
	return broker.mp[name]
}

func handlePut(w http.ResponseWriter, r *http.Request) {
	queueName := r.URL.Path[1:]
	message := r.URL.Query().Get("v")
	if message == "" {
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	broker.getQueue(queueName).Enqueue(message)
	w.WriteHeader(http.StatusOK)
}

func handleGet(w http.ResponseWriter, r *http.Request) {
	queueName := r.URL.Path[1:]
	timeoutStr := r.URL.Query().Get("timeout")
	var timeout time.Duration

	if timeoutStr != "" {
		seconds, err := strconv.Atoi(timeoutStr)
		if err != nil {
			http.Error(w, "Invalid timeout", http.StatusBadRequest)
			return
		}
		timeout = time.Duration(seconds) * time.Second
	}

	msg, ok := broker.getQueue(queueName).Dequeue(timeout)
	if !ok {
		http.Error(w, "", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprint(w, msg)
}

var broker = NewBroker()

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <port>")
		return
	}
	port := os.Args[1]

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			handlePut(w, r)
		case http.MethodGet:
			handleGet(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	fmt.Printf("Server is running on port %s\n", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Println("Error starting server: ", err)
	}
}
