package main

import (
	"dependor/lib/dependencygraph"
	"fmt"
)

func main() {
	graph := dependencygraph.NewSync("./lib/dependencygraph")
	edges, err := graph.ParseGraph()
	if err != nil {
		fmt.Printf("Got an error. Error: %s", err)
		return
	}
	printGraph(edges)
}

func printGraph(graph map[string][]string) {
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
