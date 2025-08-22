package queue

import (
	"errors"
	"fmt"
	"sync"
)

type Queue struct {
	queueArray []any
	size       int
	mu         sync.Mutex
}

func New() *Queue {
	return &Queue{queueArray: make([]any, 0), size: 0}
}

func (queue *Queue) Enqueue(value any) {
	queue.mu.Lock()
	queue.queueArray = append(queue.queueArray, value)
	queue.size = queue.size + 1
	queue.mu.Unlock()
}

func (queue *Queue) Dequeue() (any, error) {
	queue.mu.Lock()
	if queue.size == 0 {
		return nil, errors.New("queue is empty. Nothing to return")
	}
	value := queue.queueArray[0]
	queue.queueArray = queue.queueArray[1:queue.size]
	queue.size = queue.size - 1
	queue.mu.Unlock()
	return value, nil
}

func (queue *Queue) Size() int {
	return queue.size
}

func (queue *Queue) Print() {
	for _, value := range queue.queueArray {
		fmt.Print(value, ", ")
	}
	fmt.Print("\n")
}
