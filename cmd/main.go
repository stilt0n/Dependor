package main

import (
	"fmt"

	"github.com/stilt0n/dependor/lib/dependencygraph"
	"github.com/stilt0n/dependor/lib/utils"
)

func main() {
	graph := dependencygraph.NewSync("./lib/dependencygraph")
	edges, err := graph.ParseGraph()
	if err != nil {
		fmt.Printf("Got an error. Error: %s", err)
		return
	}
	printGraph(edges)
	utils.WriteGraph(edges)
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
