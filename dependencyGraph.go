package dependor

import (
	"encoding/json"
	"os"

	"github.com/stilt0n/dependor/internal/utils"
)

// An adjacency list representation of a projects imports and exports
type DependencyGraph map[string][]string

// Writes graph to a JSON file with an optional filename argument
// if no argument is provided the file will be named "dependor-output.json"
// will panic if unable to marshal json or write to the file name
func (dg DependencyGraph) WriteToJSONFile(fileName ...string) {
	writePath := "dependor-output.json"
	if len(fileName) > 0 {
		writePath = fileName[0]
	}
	asJson, err := json.Marshal(dg)
	if err != nil {
		panic(err)
	}
	err = os.WriteFile(writePath, asJson, 0666)
	if err != nil {
		panic(err)
	}
}

// Returns a JSON string of the graph. If there is an error marshaling
// JSON then the error is returned.
func (dg DependencyGraph) WriteToJSONString() (string, error) {
	jsonBytes, err := json.Marshal(dg)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

// Returns a new dependency graph with edges reversed
func (dg DependencyGraph) ReverseEdges() DependencyGraph {
	reversed := make(DependencyGraph, 0)
	for node, edges := range dg {
		for _, edge := range edges {
			reversed[edge] = append(reversed[edge], node)
		}
	}
	return reversed
}

// Performs a breadth-first traversal of the dependency graph
// starting from `startingNode` and call `fn` on each visited node
func (dg DependencyGraph) Traverse(startingNode string, fn func(node string)) {
	workQueue := utils.NewQueue[string]()
	seen := make(utils.Set[string], 0)
	workQueue.Enqueue(startingNode)
	seen.Add(startingNode)
	for !workQueue.Empty() {
		currentNode := workQueue.Dequeue()
		edges := dg[currentNode]
		fn(currentNode)
		for _, edge := range edges {
			if !seen.Has(edge) {
				workQueue.Enqueue(edge)
				seen.Add(edge)
			}
		}
	}
}
