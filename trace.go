package katana

import (
	"strings"
)

// Trace keeps track of the current dependency graph under resolution watching out for
// cyclic dependencies.
type Trace struct {
	Types []string
}

// NewTrace creates a new instance of Trace
func NewTrace() *Trace {
	return &Trace{}
}

// Empty returns true in case the trace is empty, false otherwise
func (stack *Trace) Empty() bool {
	return len(stack.Types) == 0
}

// Contains returns true in case the given typ is already in the trace, false otherwise
func (stack *Trace) Contains(typ string) bool {
	for _, t := range stack.Types {
		if t == typ {
			return true
		}
	}
	return false
}

func (stack *Trace) push(typ string) *Trace {
	stack.Types = append(stack.Types, typ)
	return stack
}

// Pop removes the last item from the trace, returning it as a result
func (stack *Trace) Pop() string {
	if len(stack.Types) == 0 {
		return ""
	}

	last := len(stack.Types) - 1
	typ := stack.Types[last]
	stack.Types = stack.Types[:last]
	return typ
}

// Push appends to the trace the given type under resolution.
// Returns a ErrCyclicDependency in case the type is already in the trace, returns nil otherwise.
func (trace *Trace) Push(typ string) (err error) {
	if trace.Contains(typ) {
		err = ErrCyclicDependency{trace}
	}
	trace.push(typ)
	return err
}

// String pretty prints the trace
func (trace *Trace) String() string {
	return "[" + strings.Join(trace.Types, " -> ") + "]"
}
