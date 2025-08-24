package linkedlist

type Queue interface {
	Enqueue(value any)
	Dequeue() (any, error)
	PeekQueue() (any, error)
	Size() int
	PrintQueue()
}

func NewQueue() Queue {
	return &NodeList{size: 0}
}

func (queue *NodeList) Enqueue(value any) {
	queue.Insert(value)
}

func (queue *NodeList) Dequeue() (any, error) {
	return queue.RemoveFirst()
}

func (queue *NodeList) PeekQueue() (any, error) {
	return queue.GetFirst()
}

func (queue NodeList) PrintQueue() {
	queue.Print()
}
