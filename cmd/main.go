package main

import (
	"flag"
	"fmt"

	"github.com/stilt0n/dependor"
)

func main() {
	var writeFlag = flag.Bool("write", false, "Write output to dependor-output.json file")
	var prettyPrintFlag = flag.Bool("pretty", false, "Pretty print output to stdout")
	flag.Parse()

	graphParser := dependor.NewSync(".")
	graph, err := graphParser.ParseGraph()
	if err != nil {
		fmt.Printf("Got an error. Error: %s", err)
		return
	}

	if *writeFlag {
		graph.WriteToJSONFile()
		return
	}

	if *prettyPrintFlag {
		printGraph(graph)
		return
	}

	jsonOutput, err := graph.WriteToJSONString()
	if err != nil {
		fmt.Printf("An error occurred when stringifying the output:\n%s", err)
	}
	fmt.Println(jsonOutput)
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
