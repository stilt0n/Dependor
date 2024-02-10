package main

import (
	"dependor/lib/dependencygraph"
	"fmt"
)

func main() {
	graph := dependencygraph.NewWithRootPath("./lib/dependencygraph")
	graph.PrintPaths()
	walked, err := graph.Walk()
	if err != nil {
		fmt.Printf("Got an error. Error: %s", err)
		return
	}
	printGraph(walked)
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
