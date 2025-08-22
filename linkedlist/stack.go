package linkedlist

type Stack interface {
	Push(value any)
	Pop() (any, error)
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

func (stack NodeList) PrintStack() {
	stack.PrintBackwards()
}
