package utils

func ReverseEdges(graph map[string][]string) map[string][]string {
	reversed := make(map[string][]string, 0)
	for node, edges := range graph {
		for _, edge := range edges {
			reversed[edge] = append(reversed[edge], node)
		}
	}
	return reversed
}

// Does a Breadth-First traversal of a graph and calls fn on each node
func TraverseFn(graph map[string][]string, startingNode string, fn func(node string)) {
	workQueue := make(Deque[string], 0)
	seen := make(Set[string], 0)
	workQueue.Enqueue(startingNode)
	for len(workQueue) > 0 {
		currentNode := workQueue.Dequeue()
		seen.Add(currentNode)
		edges := graph[currentNode]
		fn(currentNode)
		for _, edge := range edges {
			workQueue.Enqueue(edge)
		}
	}
}

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

type Deque[T any] []T

func (q Deque[T]) Enqueue(key ...T) {
	q = prepend(q, key...)
}

func (q Deque[T]) Dequeue() T {
	q, out := pop(q)
	return out
}

func (s Deque[T]) Push(key ...T) {
	s = append(s, key...)
}

// This and Dequeue are the same function because Enqueue prepends and Push appends
func (s Deque[T]) Pop() T {
	s, out := pop(s)
	return out
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
