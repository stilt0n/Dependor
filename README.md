# Dependor

A configurable JavaScript dependency graph parser written in Go.

## Why?

I have CI scripts that make use of tools that analyze the project's dependency graph to perform certain actions, but I don't have enough freedom to configure those tools to fit all of my use cases.

Having a JavaScript depency parser means I can start writing my own dependency tools.

Dependor is written in Go, which means it can be compiled and (eventually) make use of concurrency. Dependor also _only_ parses dependencies insteaded of a full JavaScript AST. This can help prevent dependency parsing from becoming too much of a bottleneck for tooling that uses it as a first step.

## How to use Dependor

### Installing Dependor

You can install dependor using:

```sh
go get github.com/stilt0n/dependor
```

Or by importing it manually and using:

```sh
go mod tidy
```

### Configuring dependor

Dependor uses a `dependor.json` file for configuration. There are two ways you can currently customize dependor:

- Add ignore glob patterns for ignoring files and directories (it is usually a good idea to ignore node_modules and build/dist directories)
- Path aliases in case you project uses any (e.g. Remix uses `~` for the `app` directory)

The `dependor.json` looks like this:

```json
{
  "ignorePatterns": ["**/node_modules", "**/dist", "**/build"],
  "pathAliases": { "~": "app" }
}
```

### Simple Example

It's easy to get started parsing dependencies with dependor:

```go
parser := dependor.NewSync(".")
graph, err := parser.ParseGraph()
if err != nil {
  return err
}
for file, imports := range graph {
  for _, imp := range imports {
    fmt.Printf("%q imports %q\n", file, imp)
  }
}
```

### Caveats

Dependor does not handle _all_ possible export syntax yet. Dependor tries to handle as many cases from the mdn docs as possible (see mdn docs for [imports](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Statements/import) and [exports](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Statements/export)) but for exports there are some cases that are not yet handled:

```js
// 1.
export let x, y;
// 2.
export const a = "a",
  b = "b",
  c = "c";
// 3.
export { foo as "invalid identifier alias" } from "./foo";
```

Cases 1 and 3 will likely be handled sometime in the near future (with lower priority on 3. which seems obscure enough that even my ESLint config thinks it's wrong). Case 2 is unlikely to be handled by Dependor any time soon, because I have been unable to think of a way to do so without implementing JavaScript expression parsing, which would pretty much require me to write a full JavaScript parser.

There is also a [known bug](https://github.com/stilt0n/dependor/issues/19) where import statements inside JSX tags are not ignored. This should cause the tokenizer to panic, so if you're not getting errors this bug probably doesn't effect you.

## Parser Methods

#### `NewSync`

Constructor for `SingleThreadedGraphParser`

**Arguments:**
`rootPath string (optional)`:

- An optional argument to tell dependor which directory to parse. If omitted dependor will parse the directory it is called from.

**Returns:**

`*SingleThreadedGraphParser`:

- A pointer to a `SingleThreadedGraphParser` struct

**Example:**

```go
parser := dependor.NewSync("./path/to/root")
```

#### `SingleThreadedGraphParser.ParseGraph()`

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

- `DependencyGraph`
  - An adjacency list representation of the parsed directory's dependencies
  - Is an alias for `map[string][]string` and can be used the same way
  - keys refer to files
  - values are lists of the files the key files import
- `error` non-nil when something goes wrong with parsing

**Example:**

```go
parser := dependor.NewSync("./path/to/root")
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

#### `SingleThreadedGraphParser.GetCustomConfig`

Retrieves custom config values from `dependor.json`. Dependor is intended to be used in other tooling and in some cases it may be useful for that tooling to piggyback on the `dependor.json` config file rather than requiring an additional config file. Dependor will parse arbitrary config values and can return values it does not make use of for other tooling to use.

**Arguments:**

None

**Returns:**

- `[]bytes` An array of marshalled JSON bytes which can be unmarshalled into an arbitrary struct
- `error` An error if there is an issue converting into JSON

**Example:**

```go
parser := dependor.NewSync()
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

### DependencyGraph Methods

The dependency graph is an alias for `map[string][]string` with some helpful receiver methods attached. Since it is just a `map` alias, it can be used the same way a map is used:

```go
// Get value using key
dependencies := dependencyGraph["foo.js"]
// Set value using key (note: you shouldn't usually modify the graph)
dependencyGraph["foo.js"] = []string{"bar.js", "baz.js"}
```

#### `WriteToJSONFile`

Writes the dependency graph to a JSON file.

**Arguments:**

`fileName string (optional)`:

- The name of the file to write to.
- If no name is provided the name `"dependor-output.json"` is used.
- Panics if it fails to marshalling json or write the file

**Returns:**

void

#### `WriteToJSONString`

Writes the dependency graph to a JSON string.

**Arguments:**

None

**Returns:**

`(string, error)`

- stringified JSON representation of the graph
- error if json marshalling fails

**Example:**

```go
json, err := graph.WriteToJSONString()
if err != nil {
  return err
}
fmt.Println(json)
```

#### `ReverseEdges`

Returns a new dependency graph with the edge directions reversed. i.e.

```go
example := DependencyGraph{ "foo": { "bar", "baz" }}
reversed := example.ReverseEdges()
// { "bar": {"foo"}, "baz": {"foo"}}
```

This can be useful if you need to figure out where a certain file is imported. _Does not modify original graph_.

**Arguments:**

None

**Returns:**

`DependencyGraph`

- A dependency graph with edges in reverse direction of the calling graph

#### `Traverse`

Performs a breadth-first traversal of the dependency graph starting from a given node and call a function on each visited node.

**Arguments:**

- `startingNode string` the node to start the traversal from
- `fn func(node string)` a function to call on each visited node

**Returns:**

void

**Example:**

```go
var indirectDependencies []string
graph.Traverse("foo.js", func(node string) {
  // these are direct dependencies
  if node != "foo.js" && !slices.Contains(graph["foo.js"], node) {
    indirectDependencies = append(indirectDependencies, node)
  }
})
fmt.Println("Indirect dependencies of foo.js:")
for _, dep := range indirectDependencies {
  fmt.Println(dep)
}
```

### Extending `dependor.json` config

`dependor.json` files can be extended to fit the use case of tooling that makes use of Dependor. See [GetCustomConfig](#getcustomconfig)
