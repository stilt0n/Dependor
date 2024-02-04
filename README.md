# Dependor

An in progress JavaScript dependency graph parser written in Go.

## Why?

I have CI scripts that use analyze JavaScript dependencies and the tools in the JavaScript ecosystem are not so fast when run on large mono-repos and in some cases not as configurable as I would like.

Dependor parses a project's dependency graph in Go which means that it can:

- Take advantage of a fast compiled language
- Use concurrency when performing file i/o
