package stack

import (
	"errors"
	"fmt"
	"sync"
)

type Stack struct {
	stackArray []any
	size       int
	mu         sync.Mutex
}

func New() *Stack {
	return &Stack{stackArray: make([]any, 0), size: 0}
}

func (stack *Stack) Push(value any) {
	stack.mu.Lock()
	defer stack.mu.Unlock()
	stack.stackArray = append(stack.stackArray, value)
	stack.size = stack.size + 1
}

func (stack *Stack) Pop() (any, error) {
	stack.mu.Lock()
	defer stack.mu.Unlock()
	if stack.size == 0 {
		return nil, errors.New("empty stack")
	}
	value := stack.stackArray[stack.size-1]
	stack.size = stack.size - 1
	stack.stackArray = stack.stackArray[:stack.size]
	return value, nil
}

func (stack Stack) Size() int {
	return stack.size
}

func (stack Stack) Print() {
	stack.mu.Lock()
	defer stack.mu.Unlock()
	if stack.size == 0 {
		fmt.Println("empty stack")
		return
	}
	for _, value := range stack.stackArray {
		fmt.Print(value, ", ")
	}
	fmt.Print("\n")
}
