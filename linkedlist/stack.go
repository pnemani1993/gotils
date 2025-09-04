package linkedlist

type Stack interface {
	Push(value any)
	Pop() (any, error)
	PeekStack() (any, error)
	Size() int
	PrintStack()
}

func NewStack() Stack {
	return &NodeList{size: 0}
}

func (stack *NodeList) Push(value any) {
	stack.Insert(value)
}

func (stack *NodeList) Pop() (any, error) {
	value, err := stack.RemoveLast()
	if err != nil {
		return nil, &InvalidOperation{1001, "empty queue"}
	}
	return value, nil
}

func (stack *NodeList) PeekStack() (any, error) {
	value, err := stack.GetLast()
	if err != nil {
		return nil, &InvalidOperation{1001, "empty queue"}
	}
	return value, nil
}

func (stack *NodeList) PrintStack() {
	stack.PrintBackwards()
}
