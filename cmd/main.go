package main

import (
	"dependor/lib/dependencygraph"
)

func main() {
	graph := dependencygraph.NewWithRootPath("./lib/dependencygraph")
	graph.PrintPaths()
}
