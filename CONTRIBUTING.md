# Contributing

There isn't an established process for contributing to dependor yet, but you can open an issue to report bugs / request features and open a PR if you'd like to add something yourself.

## Project structure

### Public API

All of dependor's public API is in the root directory. There are two parts to the public API:

- The dependency graph parser, which is located in `graphParserSync.go` (there will eventually be a concurrent parser).
- The dependency graph methods, which are located in `dependencyGraph.go`

Then there are the associated tests. The parser tests make use of the `test_tree` directory which is full of JavaScript / TypeScript files that can be parsed for testing. When adding new features or fixing bugs, it may become necessary to add files to the test tree.

For a more in depth overview of how parsing works see [How dependency parsing works](#how-dependency-parsing-works)

### Internal packages

There are a few internal packages that dependor uses when parsing:

- tokenizer
- config
- utils

The most complex of all of these package is tokenizer, which tokenizes JavaScript (or TypeScript) files. The tokenizer has its own [README](./internal/tokenizer/README.md) that goes into more detail on how tokenization works. The parser then does extra work on the file tokens to turn them into a dependency graph.

The purpose of `config` is to read `dependor.json` config files. It also has a few utility methods that make use of config information to do things like resolve paths.

`utils` has simple implementations of a Set and a Queue. These are useful for the dependency graph methods that use breadth-first search.

## How dependency parsing works

The main steps for parsing the file tree are in the `ParseGraph()` method. What parse graph does at a high level:

### Goal

The end goal is to turn a file tree that has JavaScript (or TypeScript) files and turn it into an graph that maps files to the files they import.

The end result is stored in an edge list like this:

```go
type FilePath string
type ImportPath string
type EdgeList map[FilePath][]ImportPath
```

Note that the aliases are added here for clarity and not used in the actual typing of the Edge List. The actual return type is a `DependencyGraph` which is just an alias for `map[string][]string` with a few useful receiver methods.

#### Walk filetree and tokenize files

`graph.walk()` walks the files tree and looks for files with a JavaScript related extension (at the time being this is `js|jsx|ts|tsx`). When it encounters such a file it reads it and turns it into a FileToken:

```go
type FileToken struct {
	FilePath    string
	Imports     map[string][]string
	ReExports   []string
	Exports     []string
	ReExportMap map[string]string
}
```

The file tokens are stored in a map of type:

```go
map[string]*tokenizer.FileToken
```

The key for this map will always be `FileToken.FilePath`. The reason for storing the tokens in a map will become apparent in other methods.

To see more details on FileTokens see the [tokenizer README](./internal/tokenizer/README.md).

One important detail is that the ReExportMap is only partially populated by the tokenizer. This is because some parts of creating it require all the files in the tree to be tokenized.

#### Resolve import extensions

When files are tokenized, relative paths are converted to absolute paths (w.r.t to the repo root) but JavaScript imports are not required to use file extensions:

```js
// Relevant file could be .js, .ts, .jsx, .tsx, etc.
import { foo } from "./foo";
```

File extensions are not dealt with in the tokenizer because doing it there would require extra file i/o which is expensive. Since each file path is already stored with its import/export info, we have effectively already cached the relevant parts of the file system when we tokenized it. So to figure out the extension of an extensionless import we just need to check if `extensionlessImport + extension` exists in the token map. If it does, then the path will be resolved to use that extension. For imports such as named imports (e.g. `import React from 'react';`) no extension will be added.

Import aliases are also handled here. These could potentially be handled in the tokenizer in the future, but were more convenient to handle at parse time with how things are currently structured.

#### Finish Index Maps

ES Imports have a cool, but challenging to parse, feature: you can import from a directory that has an `index.js` file in it:

```js
// In foo.js
import { bar } from "./components/Bar";

// in components/Bar/index.js
export * from "./bar";
export * from "./baz";
```

To figure out where `bar` is _really_ coming from, we need to be able to track which identifiers are exported out of `index.js` and which files those identifiers belong to. We also need to know which identifiers were imported inside of `foo.js`.

Having a ReExport map allows us to handle this. A reExport map looks like this:

```go
reExports := map[string]string {
  "identifier": "exportedFrom",
  "bar": "./bar.js",
  "useBar": "./bar.js",
  "baz": "./baz.js",
}
```

This is the final piece of setup before we are ready to turn the tokens into an edge list.

#### Parse Tokens

The `parseTokens` method goes through all of the updated FileTokens and appends their path: exports to the graph. The code for this is pretty simple. In pseudocode:

```py
edgeList = {}
for token in tokens:
  for path, identifiers in token.Imports:
    edges = []
    if isIndexFile(path):
      edges.extend(resolveIndexImport(path, identifiers))
    else:
      edges.append(edges, path)
  edgeList[token.FilePath] = edges
```

This is pretty much the exact code in the file. `resolveIndexImport` uses the identifiers that are being imported and the Index Map we created to figure out which files are being imported from. This can be multiple files, for example in the last example if we imported:

```js
import { bar, baz } from "./components/Bar";
```

We would be importing from `components/Bar/bar.js` and `components/Bar/baz.js`.

When this step is finished. The edge list is returned.

## Adding a concurrent api

Many of these steps could benefit from concurrency. I am trying to keep this in mind in how I structure the steps so that they can be made concurrent in the future. But to have a good baseline, I am starting with a single-threaded api.
