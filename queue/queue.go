package queue

import (
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
	defer queue.mu.Unlock()
	queue.queueArray = append(queue.queueArray, value)
	queue.size = queue.size + 1
}

func (queue *Queue) Dequeue() (any, error) {
	queue.mu.Lock()
	defer queue.mu.Unlock()
	if queue.size == 0 {
		return nil, &InvalidOperation{1001, "empty queue"}
	}
	value := queue.queueArray[0]
	queue.queueArray = queue.queueArray[1:queue.size]
	queue.size = queue.size - 1
	return value, nil
}

func (queue *Queue) Size() int {
	return queue.size
}

func (queue *Queue) Print() {
	queue.mu.Lock()
	defer queue.mu.Unlock()
	if queue.size == 0 {
		fmt.Println("empty queue")
		return
	}
	for _, value := range queue.queueArray {
		fmt.Print(value, ", ")
	}
	fmt.Print("\n")
}

type InvalidOperation struct {
	code    int
	message string
}

func (err *InvalidOperation) Error() string {
	return fmt.Sprintf("Error %d: %s", err.code, err.message)
}
