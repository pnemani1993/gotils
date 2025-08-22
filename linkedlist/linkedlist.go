package linkedlist

import (
	"errors"
	"fmt"
	"sync"
)

type Node struct {
	value    any
	next     *Node
	previous *Node
}

type LinkedList interface {
	Insert(value any)
	RemoveFirst() (any, error)
	RemoveLast() (any, error)
	Remove(index int) (any, error)
	Size() int
	Get(index int) (any, error)
	GetFirst() (any, error)
	GetLast() (any, error)
	Print()
}

type NodeList struct {
	head *Node
	tail *Node
	size int
	mu   sync.Mutex
}

func New() LinkedList {
	return &NodeList{size: 0}
}

func (dll *NodeList) Insert(value any) {
	newNode := &Node{value: value}
	dll.mu.Lock()
	if dll.size == 0 {
		dll.head = newNode
		dll.tail = newNode
		dll.size = 1
		dll.mu.Unlock()
		return
	}
	dll.tail.next = newNode
	newNode.previous = dll.tail
	dll.tail = newNode
	dll.size = dll.size + 1
	dll.mu.Unlock()
}

func (dll *NodeList) RemoveFirst() (any, error) {
	if dll.size == 0 {
		return nil, errors.New("linkedList is empty. Nothing to remove")
	}
	dll.mu.Lock()
	node := dll.head
	returnValue := node.value
	if dll.size == 1 {
		dll.size = 0
		node = nil
		dll.mu.Unlock()
		return returnValue, nil
	}
	dll.head = dll.head.next
	dll.head.previous = nil
	node.next = nil
	node = nil
	dll.size = dll.size - 1
	dll.mu.Unlock()
	return returnValue, nil
}

func (dll *NodeList) RemoveLast() (any, error) {
	if dll.size == 0 {
		return nil, errors.New("linkedList is empty. Nothing to remove")
	}
	dll.mu.Lock()
	node := dll.tail
	returnValue := node.value
	dll.tail = dll.tail.previous
	if dll.size == 1 {
		dll.size = 0
		node = nil
		dll.mu.Unlock()
		return returnValue, nil
	}
	dll.tail.next = nil
	node.previous = nil
	node = nil
	dll.size = dll.size - 1
	dll.mu.Unlock()
	return returnValue, nil
}

func (dll *NodeList) Remove(index int) (any, error) {
	if dll.size == 0 || index >= dll.size || index < 0 {
		return nil, errors.New("Invalid index provided or the list is empty.")
	}
	if index == 0 {
		return dll.RemoveFirst()
	}
	if index == dll.size-1 {
		return dll.RemoveLast()
	}

	dll.mu.Lock()

	node := dll.head
	for i := 0; i < index; i = i + 1 {
		node = node.next
	}
	returnValue := node.value
	prevNode := node.previous
	nextNode := node.next

	// severing ties with the surrounding nodes
	prevNode.next = nextNode
	nextNode.previous = prevNode

	// nullifying the node and storing the value in another variable
	node.next = nil
	node.previous = nil
	node = &Node{}
	dll.size = dll.size - 1
	dll.mu.Unlock()
	return returnValue, nil
}

func (dll NodeList) Size() int {
	return dll.size
}

func (dll NodeList) Get(index int) (any, error) {
	if dll.size == 0 || index >= dll.size || index < 0 {
		return nil, errors.New("Invalid index provided or the list is empty.")
	}
	dll.mu.Lock()
	if index == 0 {
		returnValue := dll.head.value
		dll.mu.Unlock()
		return returnValue, nil
	}
	if index == dll.size-1 {
		returnValue := dll.tail.value
		dll.mu.Unlock()
		return returnValue, nil
	}
	node := dll.head
	for i := 0; i < index; i = i + 1 {
		node = node.next
	}
	returnValue := node.value
	dll.mu.Unlock()
	return returnValue, nil
}

func (dll NodeList) GetFirst() (any, error) {
	return dll.Get(0)
}

func (dll NodeList) GetLast() (any, error) {
	return dll.Get(dll.Size() - 1)
}

func (dll NodeList) Print() {
	dll.mu.Lock()
	if dll.size == 0 {
		fmt.Print("The linkedlist is empty")
		dll.mu.Unlock()
		return
	}
	if dll.size == 1 {
		fmt.Print(dll.head.value, "\n")
		dll.mu.Unlock()
		return
	}
	node := dll.head
	for i := 0; i < dll.size; i = i + 1 {
		fmt.Print(node.value, ", ")
		node = node.next
	}
	fmt.Print("\n")
	dll.mu.Unlock()
}

func (dll NodeList) PrintBackwards() {
	dll.mu.Lock()
	if dll.size == 0 {
		fmt.Print("The linkedlist is empty")
		dll.mu.Unlock()
		return
	}
	if dll.size == 1 {
		fmt.Print(dll.head.value, "\n")
		dll.mu.Unlock()
		return
	}
	node := dll.tail
	for i := 0; i < dll.size; i = i + 1 {
		fmt.Print(node.value, ", ")
		node = node.previous
	}
	fmt.Print("\n")
	dll.mu.Unlock()
}
