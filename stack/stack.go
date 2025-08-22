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
	stack.stackArray = append(stack.stackArray, value)
	stack.size = stack.size + 1
	stack.mu.Unlock()
}

func (stack *Stack) Pop() (any, error) {
	stack.mu.Lock()
	if stack.size == 0 {
		return nil, errors.New("stack is empty. nothing to return")
	}
	value := stack.stackArray[stack.size-1]
	stack.size = stack.size - 1
	stack.stackArray = stack.stackArray[:stack.size]
	stack.mu.Unlock()
	return value, nil
}

func (stack Stack) Size() int {
	return stack.size
}

func (stack Stack) Print() {
	stack.mu.Lock()
	for _, value := range stack.stackArray {
		fmt.Print(value, ", ")
	}
	fmt.Print("\n")
	stack.mu.Unlock()
}
