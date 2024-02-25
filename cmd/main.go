package main

import (
	"fmt"

	"github.com/stilt0n/dependor"
)

func main() {
	graphParser := dependor.NewSync(".")
	graph, err := graphParser.ParseGraph()
	if err != nil {
		fmt.Printf("Got an error. Error: %s", err)
		return
	}
	printGraph(graph)
	graph.WriteToJSONFile()
}

func printGraph(graph dependor.DependencyGraph) {
	for node, edges := range graph {
		fmt.Printf("%q: {", node)
		if len(edges) > 0 {
			fmt.Print("\n")
		}
		for _, edge := range edges {
			fmt.Printf("  %q,\n", edge)
		}
		fmt.Println("}")
	}
}
