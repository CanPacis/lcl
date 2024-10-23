package internal

type Stack[T any] struct {
	values []T
}

func (s *Stack[T]) Push(v T) {
	s.values = append(s.values, v)
}

func (s *Stack[T]) Pop() T {
	if len(s.values) == 0 {
		var t T
		return t
	}

	popped := s.values[len(s.values)-1]
	s.values = s.values[:len(s.values)-1]
	return popped
}

func (s Stack[T]) Last() T {
	if len(s.values) == 0 {
		var t T
		return t
	}

	return s.values[len(s.values)-1]
}

func NewStack[T any](values ...T) *Stack[T] {
	if values == nil {
		values = []T{}
	}
	return &Stack[T]{values: values}
}
