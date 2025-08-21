package linkedlist

import (
	"fmt"
	"log"
	"sync"
)

type Node struct {
	value    any
	next     *Node
	previous *Node
}

type LinkedList interface {
	Insert(value any)
	RemoveFirst() any
	RemoveLast() any
	Remove(index int) any
	Size() int
	Get(index int) any
	GetFirst() any
	GetLast() any
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

func (dll *NodeList) RemoveFirst() any {
	if dll.size == 0 {
		log.Fatal("The LinkedList is empty. Nothing to remove")
	}
	dll.mu.Lock()
	returnValue := dll.head
	dll.head = dll.head.next
	if dll.size == 1 {
		dll.size = 0
		dll.mu.Unlock()
		return returnValue
	}
	dll.head.previous = nil
	returnValue.next = nil
	dll.size = dll.size - 1
	dll.mu.Unlock()
	return returnValue
}

func (dll *NodeList) RemoveLast() any {
	if dll.size == 0 {
		log.Fatal("The LinkedList is empty. Nothing to remove")
	}
	dll.mu.Lock()
	returnValue := dll.tail
	dll.tail = dll.tail.previous
	if dll.size == 1 {
		dll.size = 0
		dll.mu.Unlock()
		return returnValue
	}
	dll.tail.next = nil
	returnValue.previous = nil
	dll.size = dll.size - 1
	dll.mu.Unlock()
	return returnValue
}

func (dll *NodeList) Remove(index int) any {
	return 0
}
func (dll NodeList) Size() int {
	return dll.size
}
func (dll NodeList) Get(index int) any {
	if dll.size == 0 || index >= dll.size || index < 0 {
		log.Fatal("Invalid index provided or the list is empty.")
	}
	dll.mu.Lock()
	if index == 0 {
		returnValue := dll.head.value
		dll.mu.Unlock()
		return returnValue
	}
	if index == dll.size-1 {
		returnValue := dll.tail.value
		dll.mu.Unlock()
		return returnValue
	}
	node := dll.head
	for i := 0; i < index; i = i + 1 {
		node = node.next
	}
	returnValue := node.value
	dll.mu.Unlock()
	return returnValue
}
func (dll NodeList) GetFirst() any {
	return dll.Get(0)
}
func (dll NodeList) GetLast() any {
	return dll.Get(dll.Size() - 1)
}

func (dll NodeList) Print() {
	if dll.size == 0 {
		fmt.Print("The linkedlist is empty")
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
