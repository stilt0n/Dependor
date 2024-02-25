package utils

func prepend[T any](keys []T, key ...T) []T {
	return append(key, keys...)
}

func pop[T any](keys []T) ([]T, T) {
	var defaultValue T
	if len(keys) == 0 {
		return []T{}, defaultValue
	}
	poppedKey := keys[len(keys)-1]
	return keys[:len(keys)-1], poppedKey
}

type Queue[T any] struct {
	queue []T
}

func NewQueue[T any]() *Queue[T] {
	return &Queue[T]{make([]T, 0)}
}

func (q *Queue[T]) Enqueue(key ...T) {
	q.queue = prepend(q.queue, key...)
}

func (q *Queue[T]) Dequeue() T {
	updated, out := pop(q.queue)
	q.queue = updated
	return out
}

func (q *Queue[T]) Empty() bool {
	return len(q.queue) == 0
}

func (q *Queue[T]) Length() int {
	return len(q.queue)
}

type Set[T comparable] map[T]bool

func (s Set[T]) Add(key T) {
	s[key] = true
}

func (s Set[T]) Has(key T) bool {
	_, ok := s[key]
	return ok
}

func (s Set[T]) Keys() []T {
	keys := make([]T, len(s))
	i := 0
	for k := range s {
		keys[i] = k
		i++
	}
	return keys
}
