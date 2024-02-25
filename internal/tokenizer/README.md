# Overview

This package converts JavaScript (or TypeScript) into tokens which can be used when parsing a file tree into a dependency graph.

## How it works

The tokenizer reads each character one at a time and finds:

- imported paths
- re-exported paths
- imported tokens
- exported tokens

Imported identifiers need to be associated with their paths, so they stored in a map:

```go
// In the code these are just unaliased strings
// but here the aliases are added for clarity
type importPath string
type importedIdentifier string
type Imports map[importPath][]importedIdentifier
```

This makes the type for file token:

```go
type fileToken struct {
	FilePath  string
	Imports   map[string][]string
  ReExports []string
	Exports   []string
}
```

### Finding imports

For dynamic imports and require statements only the import paths are tracked because additional information is unnecessary to resolve those paths.

For esmodule imports:

- Every identifier between the keyword `import` and the keyword `from` is an identifier
- The first string after `import` is the import path
- Ignore identifiers after keyword `as`
- Identifiers are separated by stop characters which are:
  - Whitespace
  - Commas
  - Curly Braces
  - Slash (i.e. comment starts)

The tokenizer assumes that it is being given valid JavaScript syntax. Syntax errrors may cause issues with tokenization. I have no plans to address this unless it can be done in a way that doesn't degrade performance.

### Exports

I am not tracking common js exports for now since I only track exports to allow me correctly route re-exports at parse time. I don't think you can re-export in common js and I am generally assuming that people are not mixing es exports and common js require statements in a way where they are re-exporting from a common js file. I also don't think you can re-export dynamic imports.

The export cases I'm currently considering are:

```js
export const foo = [...etc];
export { foo, bar, baz };
export default etc;
```

For the default export case, we just store "default" as the identifier since it can have an arbitrary name when imported.

For the other cases, we can store all non-keyword identifiers up to `=` or `;`.

We also need to deal with re-exports:

```js
export { foo, bar, baz } from "./foo";
```

Here, we need to make sure we don't treat the re-exports as exports since they are just being forwarded from another file. When we run into the `from` token, then the exported identifiers should be ignored and the following path should be saved to `reExports`.

Finally, exports can also have aliases, which need to be treated differently from import alaises. For an import, we should skip the alias to allow for correct mapping between files. But for exports, we need to use the alias and skip the unaliased identifier. To do this, we can simply use the `as` token as a flag to overwrite the previous index in the `exports` array.

### Re-exports

Since these are always preceded by the `export` token, we can handle them in the same method that handles exports. If we run into the `from` token, then we will search for the next string and use that as the re-export path.

#### Re-export identifiers

Re-export identifiers will eventually need to be stored to help connect importing files to the correct exporting files. Some of this can be taken care of in the tokenizer. In casese like this:

```js
export { default as foo, bar, baz } from "./foo";
```

We can just add the identifiers to the file token's reExportMap:

```go
{
  "foo": "./foo",
  "bar": "./foo",
  "baz": "./foo",
}
```

But for wildcard re-exports we won't be able to update the map until after tokenization is finished. (It is technically possibly but would involve unnecessary file i/o). In this case, to make the graph parser aware of the need to update the map the file path will be stored as a key with an asterisk as the value:

i.e.

```js
export * from "./foo";
```

results in:

```go
{
  "./foo": "*",
}
```

## Why a tokenizer?

I first tried to find imports using reglar expressions. Finding a standard import / re-export is pretty simple since we're just looking for anything between `from <quote-char>` and `<quote-char>`. But the ability to put comments in weird places makes this much harder to solve with regex. Go's lack of support for lookbehind further complicates this approach and since it turns out lookbehind is not supported due to performance concerns, I didn't think it made sense to look for a library to handle this.

If this was the only issue, I'd probably be content to let:

```js
import foo from /* "fake-import" */ "real-import";
```

remain a bug since I don't think people generally do this, or at the very least it's not a use-case I care a lot about.

The main challenge is that esmodules allow you to re-export without using `import`. Re-exporting is pretty nasty to deal with because you can re-export from multiple files, wildcard re-export, and then import those from a directory if they are in a file called `index.(js|ts|jsx...)`. Unfortunately, I work with code that does this:

```js
// foo.js
import { bar, baz } from "./utils";

// utils/index.js
export * from "./pathUtils";
export * from "./fileUtils";
export * from "./miscUtils";
```

Which means that I can't solve the problem of where `foo.js` is _really_ being imported from without some information about what is being exported in each file. I'd like to process this all in one pass if possible, so I've chosen to write a tokenizer for this purpose.
