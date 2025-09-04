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
	value, err := queue.RemoveFirst()
	if err != nil {
		return nil, &InvalidOperation{1001, "empty queue"}
	}
	return value, nil
}

func (queue *NodeList) PeekQueue() (any, error) {
	value, err := queue.GetFirst()
	if err != nil {
		return nil, &InvalidOperation{1001, "empty queue"}
	}
	return value, nil
}

func (queue *NodeList) PrintQueue() {
	queue.Print()
}
