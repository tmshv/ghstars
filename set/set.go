package set

type Set[T comparable] struct {
	elements map[T]bool
}

func New[T comparable]() *Set[T] {
	return &Set[T]{elements: make(map[T]bool)}
}

func (s *Set[T]) Add(value T) {
	s.elements[value] = true
}

func (s *Set[T]) Del(value T) {
	delete(s.elements, value)
}

func (s *Set[T]) Has(value T) bool {
	_, exists := s.elements[value]
	return exists
}

func (s *Set[T]) Items() []T {
	items := make([]T, 0, len(s.elements))
	for value := range s.elements {
		items = append(items, value)
	}
	return items
}
