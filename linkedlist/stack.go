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
	return stack.RemoveLast()
}

func (stack *NodeList) PeekStack() (any, error) {
	return stack.GetLast()
}

func (stack NodeList) PrintStack() {
	stack.PrintBackwards()
}
