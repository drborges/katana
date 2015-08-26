package katana

import "fmt"

type Stack struct {
	items []string
}

func NewStack() *Stack {
	return &Stack{}
}

func (stack *Stack) Empty() bool {
	return len(stack.items) == 0
}

func (stack *Stack) Reset() {
	stack.items = []string{}
}

func (stack *Stack) Contains(item string) bool {
	for _, i := range stack.items {
		if i == item {
			return true
		}
	}
	return false
}

func (stack *Stack) Push(item string) *Stack {
	stack.items = append(stack.items, item)
	return stack
}

func (stack *Stack) Pop() string {
	if len(stack.items) == 0 {
		return ""
	}

	last := len(stack.items) - 1
	item := stack.items[last]
	stack.items = stack.items[:last]
	return item
}

func (stack *Stack) String() string {
	return fmt.Sprint(stack.items)
}
