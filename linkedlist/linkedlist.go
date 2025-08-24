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
	defer dll.mu.Unlock()
	if dll.size == 0 {
		dll.head = newNode
		dll.tail = newNode
		dll.size = 1
		return
	}
	dll.tail.next = newNode
	newNode.previous = dll.tail
	dll.tail = newNode
	dll.size = dll.size + 1
}

func (dll *NodeList) RemoveFirst() (any, error) {
	if dll.size == 0 {
		return nil, errors.New("empty linkedList")
	}
	dll.mu.Lock()
	defer dll.mu.Unlock()
	node := dll.head
	returnValue := node.value
	if dll.size == 1 {
		dll.size = 0
		node = nil
		return returnValue, nil
	}
	dll.head = dll.head.next
	dll.head.previous = nil
	node.next = nil
	node = nil
	dll.size = dll.size - 1
	return returnValue, nil
}

func (dll *NodeList) RemoveLast() (any, error) {
	if dll.size == 0 {
		return nil, errors.New("empty linkedList")
	}
	dll.mu.Lock()
	defer dll.mu.Unlock()
	node := dll.tail
	returnValue := node.value
	dll.tail = dll.tail.previous
	if dll.size == 1 {
		dll.size = 0
		node = nil
		return returnValue, nil
	}
	dll.tail.next = nil
	node.previous = nil
	node = nil
	dll.size = dll.size - 1
	return returnValue, nil
}

func (dll *NodeList) Remove(index int) (any, error) {
	dll.mu.Lock()
	defer dll.mu.Unlock()
	if dll.size == 0 || index >= dll.size || index < 0 {
		if dll.size == 0 {
			return nil, errors.New("empty linkedlist")
		} else {
			return nil, errors.New("invalid index")
		}
	}
	if index == 0 {
		return dll.RemoveFirst()
	}
	if index == dll.size-1 {
		return dll.RemoveLast()
	}

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
	return returnValue, nil
}

func (dll NodeList) Size() int {
	dll.mu.Lock()
	returnSize := dll.size
	dll.mu.Unlock()
	return returnSize
}

func (dll NodeList) Get(index int) (any, error) {
	dll.mu.Lock()
	defer dll.mu.Unlock()
	if dll.size == 0 || index >= dll.size || index < 0 {
		if dll.size == 0 {
			return nil, errors.New("empty linkedlist")
		} else {
			return nil, errors.New("invalid index")
		}
	}

	if index == 0 {
		returnValue := dll.head.value
		return returnValue, nil
	}
	if index == dll.size-1 {
		returnValue := dll.tail.value
		return returnValue, nil
	}
	node := dll.head
	for i := 0; i < index; i = i + 1 {
		node = node.next
	}
	returnValue := node.value
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
	defer dll.mu.Unlock()
	if dll.size == 0 {
		fmt.Print("empty linkedlist")
		return
	}
	if dll.size == 1 {
		fmt.Print(dll.head.value, "\n")
		return
	}
	node := dll.head
	for i := 0; i < dll.size; i = i + 1 {
		fmt.Print(node.value, ", ")
		node = node.next
	}
	fmt.Print("\n")
}

func (dll NodeList) PrintBackwards() {
	dll.mu.Lock()
	defer dll.mu.Unlock()
	if dll.size == 0 {
		fmt.Println("empty linkedlist")
		return
	}
	if dll.size == 1 {
		fmt.Print(dll.head.value, "\n")
		return
	}
	node := dll.tail
	for i := 0; i < dll.size; i = i + 1 {
		fmt.Print(node.value, ", ")
		node = node.previous
	}
	fmt.Print("\n")
}
