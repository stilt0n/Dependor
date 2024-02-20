# Dependor

A configurable JavaScript dependency graph parser written in Go.

## Why?

I have CI scripts that make use of tools that analyze the project's dependency graph to perform certain actions, but I don't have enough freedom to configure those tools to fit all of my use cases.

Having a JavaScript depency parser means I can start writing my own dependency tools.

Dependor is written in Go, which means it can be compiled and (eventually) make use of concurrency. Dependor also _only_ parses dependencies insteaded of a full JavaScript AST. This can help prevent dependency parsing from becoming too much of a bottleneck for tooling that uses it as a first step.

## How to use Dependor

Dependor is still somewhat in progress. The main API is the `dependencygraph` package.

### `dependencygraph`

The `dependencygraph` package is responsible for parsing a filetree into a JavaScript dependency graph. It has three public methods.

#### `NewSync`

Constructor for `SingleThreadedGraph`
**Arguments:**
`rootPath string (optional)`:

- An optional argument to tell dependor which directory to parse. If omitted dependor will parse the directory it is called from.

**Returns:**
`*SingleThreadedGraph`:

- A pointer to a `SingleThreadedGraph` struct

**Example:**

```go
parser := NewSync("./path/to/root")
```

#### `SingleThreadedGraph.ParseGraph()`

Parses the file tree into an adjacency list representation of the file tree's JavaScript dependency structure. For example this file:

```js
// rootPath/files/foo.js
import React from "react";
import Foo, { bar, b as baz } from "./bar";
import { x, y, z } from "../letters";
import { JSXComponent } from "../components/JSXComponent";
```

Is parsed into the graph node:

```go
map[string][]string {
  // ... other nodes
  "rootPath/files/foo.js": { "react", "rootPath/files/bar.js", "rooPath/letters.js", "rootPath/components/JSXComponent.jsx" },
  // ... other nodes
}
```

**Arguments:**
None

**Returns:**

- `map[string][]string`
  - An adjacency list representation of the parsed directory's dependencies
  - keys refer to files
  - values are lists of the files those files import
- `error` non-nil when something goes wrong with parsing

**Example:**

```go
parser := NewSync("./path/to/root")
graph, err := parser.ParseGraph()
if err != nil {
  handleError(err)
}
for file, imports := range graph {
  for _, imprt := range imports {
    fmt.Printf("%q imports %q\n", file, imprt)
  }
}
```

#### `SingleThreadedGraph.GetCustomConfig`

Retrieves custom config values from `dependor.json`. Dependor is intended to be used in other tooling and in some cases it may be useful for that tooling to piggyback on the `dependor.json` config file rather than requiring an additional config file. Dependor will parse arbitrary config values and can return ones it does not make use of for other tooling to make use of.

**Arguments:**
None

**Returns:**

- `[]bytes` An array of marshalled JSON bytes which can be unmarshalled into an arbitrary struct
- `error` An error if there is an issue converting into JSON

**Example:**

```go
parser := NewSync()
graph, err := parser.ParseGraph()

// ...

type CustomConfigOptions struct {
  Foo {
    Bar:    []string `json:"bar"`,
    IsCool: bool `json:"isCool"`
  } `json:"foo"`
}

var customConfig CustomConfigOptions
jsonBytes, err := parser.GetCustomConfig()
if err != nil {
  panic(err)
}

if err := json.Unmarshal(jsonBytes, &customConfig); err == nil {
  myGraphFunc(graph, customConfig)
}
```

### `utils`

Utils has several utility functions and data structures that can be helpful for working with graphs. Utils is primarily intended for internal use and less likely to be stable than dependencygraph.

#### `Deque`

A simple implementation of a double-ended queue. Primarily intended for internal use but is generic and functions more-or-less as the method names suggest it should.

**Methods:**

- `Enqueue`
- `Dequeue`
- `Push`
- `Pop`

#### `Set`

A simple implementation of a set. Basically a wrapper for `map[T any]bool`. Primarily intended for internal use.

**Methods:**

- `Has`
- `Add`
- `Keys`

#### `WriteGraph`

Writes a graph to a json file called `dependor-output.json`.

**Arguments:**

- `graph map[string][]string` edge list representation of a graph

**Returns:**
void

**Example:**

```go
parser := dependencygraph.NewSync()
graph, err := parser.ParseGraph()
if err == nil {
  utils.WriteGraph(graph)
}
```

#### `ReverseEdges`

The graph tracks which files import other files, but you may also want to track which files are imported by other files. Reversing the graph allows you to do this. Does not modify original graph.

e.g Turns:

```go
{
  // files imported by foo.js
  "rootPath/files/foo.js": { "react", "rootPath/files/bar.js", "rootPath/letters.js", "rootPath/components/JSXComponent.jsx" },
}
```

into:

```go
{
  // files that import 'react'
  "react": {"rootPath/files/foo.js", /* ... */},
  "rooPath/files/bar.js": { "rootPath/files/foo.js" },
  "rootPath/letters.js": { "rootPath/files/foo.js" },
  "rootPath/components/JSXComponent.jsx": { "rootPath/files/foo.js" },
}
```

**Arguments:**

- `graph map[string][]string` graph to reverse (does not modify)

**Returns:**

- `map[string][]string` reversed version of `graph`

#### TraverseFn

Performs a breadth-first traversal of a graph starting from `startingNode` and calls `fn` on each node traversed

**Arguments:**

- `graph map[string][]string` graph to traverse
- `startingNode string` starting place for graph traversal
- `fn func(node string)` function to call on each node in graph

**Returns:**
void

**Example:**

```go
indirectImports := make([]string, 0)
utils.TraverseFn(graph, "rootPath/files/foo.js", func(node string) {
  indirectImports = append(indirectImports, node)
})
```

### Tokenizer

The tokenizer is intended for internal use only. See [tokenizer README](./lib/tokenizer/README.md) for details about how tokenizer works.

### Config

`dependor.json` files can be extended to fit the use case of tooling that makes use of Dependor. See [GetCustomConfig](#getcustomconfig)
