# Dependor

An in-progess JavaScript dependency graph parser written in Go.

## Why?

I have CI scripts that need to analyze JavaScript dependencies and the tools in the JavaScript ecosystem are not so fast when run on large mono-repos. I also wanted more configurability which you can get when you write things yourself.

Dependor parses a project's dependency graph in Go which means that it can:

- Take advantage of a fast compiled language
- Use concurrency when performing file i/o
