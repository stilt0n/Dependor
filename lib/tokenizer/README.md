# Overview

This package converts JavaScript (or related) into tokens that can be used to build a dependency graph.

## How it works

The tokenizer reads each character one at a time and finds:

- imported paths
- re-exported paths
- imported tokens
- exported tokens

Imported identifiers need to be associated with their paths, so they are grouped in an `importToken` struct:

```go
type importToken struct {
	importPath          string
	importedIdentifiers []string
}
```

This makes the type for file token:

```go
type fileToken struct {
	filePath  string
	imports   []importToken
  reExports []string
	exports   []string
}
```

### Finding imports

For dynamic imports and require statements I am only tracking the import paths because, at least for my current use case, that is sufficient.

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

I am not tracking common js exports for now since I only track exports to correctly route re-exports. I don't think you can re-export in common js and I am generally assuming that people are not mixing es exports and common js require statements in a way where they are re-exporting from a common js file. I am also assuming dynamic imports are not re-exported.

The export cases I'm currently considering are:

```js
export foo;
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

Here, we need to make sure we don't treat the re-exports as exports since they are just being forwarded from another file. When we run into the `from` token, then the exported identifiers should be ignored and the follwing path should be saved to `reExports`.

### Re-exports

Since these are always preceded by the `export` token, we can handle them in the same method that handles exports. If we run into the `from` token, then we will search for the next string and use that as the re-export path.

#### Re-export identifiers

We could potentially store these in the future I am choosing to treat all `export ... from 'file';` statements as if they were `export * from 'file';`. Since files track their own exports, storing a file path is sufficient to find the related exports. Since I don't currently need anything more granular than which files import from which other files I have decided it is not worth storing this information two places.

## Why a tokenizer?

I first tried to find imports using reglar expressions. Finding a standard import / re-export is pretty simple since we're just looking for anything between `from <quote-char>` and `<quote-char>`. But the ability to put comments in nasty places makes this much harder to solve with regex, and Go's lack of support for lookbehind doesn't help make this easier and since lookbehind is not supported due to performance concerns, I didn't think it made sense to look for a library to handle this.

If this was the only issue, I'd probably be content to let:

```js
import foo from /* "fake-import" */ "real-import";
```

remain a bug since I don't think people generally do this, or at the very least it's a use-case I care a lot about.

The main challenge is that esmodules allow you to re-export without using `import`. Re-exporting is pretty nasty because you can re-export from multiple files, wildcard re-export, and then import those from a directory if they are in a file called `index.(js|ts|jsx...)`. Unfortunately, I work with code that does this:

```js
// foo.js
import { bar, baz } from "./utils";

// utils/index.js
export * from "./pathUtils";
export * from "./fileUtils";
export * from "./miscUtils";
```

Which means that I can't solve the problem of where `foo.js` is _really_ being imported from without some information about what is being exported in each file. I'd like to process this all in one pass if possible, so I've chosen to write a tokenizer for this purpose.
