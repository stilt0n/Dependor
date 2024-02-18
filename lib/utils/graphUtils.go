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

func TraverseFn(graph map[string][]string, startingNode string, fn func(node string)) {
	workQueue := make(Deque, 0)
	seen := make(Set, 0)
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

func prepend(keys []string, key ...string) []string {
	return append(key, keys...)
}

func pop(keys []string) ([]string, string) {
	if len(keys) == 0 {
		return []string{}, ""
	}
	poppedKey := keys[len(keys)-1]
	return keys[:len(keys)-1], poppedKey
}

type Deque []string

func (q Deque) Enqueue(key ...string) {
	q = prepend(q, key...)
}

func (q Deque) Dequeue() string {
	q, out := pop(q)
	return out
}

func (s Deque) Push(key ...string) {
	s = append(s, key...)
}

// This and Dequeue are the same function because Enqueue prepends and Push appends
func (s Deque) Pop() string {
	s, out := pop(s)
	return out
}

type Set map[string]bool

func (s Set) Add(key string) {
	s[key] = true
}

func (s Set) Has(key string) bool {
	_, ok := s[key]
	return ok
}

func (s Set) Keys() []string {
	keys := make([]string, len(s))
	i := 0
	for k := range s {
		keys[i] = k
		i++
	}
	return keys
}
