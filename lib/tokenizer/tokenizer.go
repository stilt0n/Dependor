/*
Regex seems to mostly work for this but is a little challenging when it comes to handling comments
especially since Go does not support lookbehind regex (for performance reasons, apparently).
I didn't think it would be crazy difficult to write an import lexer so that's what this is.
*/
package tokenizer

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"unicode"
)

const (
	REQUIRE_LENGTH = 7
	IMPORT_LENGTH  = 6
	EXPORT_LENGTH  = 6
)

type Tokenizer struct {
	currentIndex int
	fileRunes    []rune
	imports      map[string][]string
	reExports    []string
	reExportMap  map[string]string
	exports      []string
	callDir      string
	initPath     string
}

func NewTokenizerFromFile(initPath string) (*Tokenizer, error) {
	file, err := os.ReadFile(initPath)
	if err != nil {
		return &Tokenizer{}, err
	}
	fileString := string(file)
	return New(fileString, initPath), nil
}

func New(fileString, initPath string) *Tokenizer {
	t := Tokenizer{
		currentIndex: 0,
		fileRunes:    []rune(fileString),
		imports:      make(map[string][]string, 0),
		reExports:    []string{},
		reExportMap:  nil,
		exports:      []string{},
		callDir:      filepath.Dir(initPath),
		initPath:     initPath,
	}
	return &t
}

// Something to keep in mind: this will at least miss *some* edge cases.
// e.g. nested requires `require(require('./pathToRealImport'))`
// Supporting this case seems challenging and I don't think it's currently worth my effort
// Reads all import paths from a file in one pass and returns them in an array
func (t *Tokenizer) Tokenize() FileToken {
	if len(t.fileRunes) < 1 {
		return FileToken{}
	}

	for t.currentIndex < t.end() {
		char := t.current()
		switch char {
		case 'i':
			t.tokenizeImport()
		case 'r':
			t.tokenizeRequire()
		case 'e':
			t.tokenizeExport()
		case '/':
			t.skipComment(t.peek())
		}
		t.advanceChars()
	}
	return FileToken{
		FilePath:    t.initPath,
		Imports:     t.imports,
		ReExports:   t.reExports,
		ReExportMap: t.reExportMap,
		Exports:     t.exports,
	}
}

func (t *Tokenizer) skipComment(peekChar rune) {
	switch peekChar {
	case '/':
		t.skipSingleLineComment()
	case '*':
		t.skipMultiLineComment()
	default:
		return
	}
}

// Finds the first quoted string after an import keyword and adds it to imports
// TODO: do I also need to handle identifiers inside of parens?
func (t *Tokenizer) tokenizeImport() {
	if t.currentIndex+IMPORT_LENGTH > t.end() {
		return
	}

	// the word `import` could be used in a variable name so this protects agains that
	if t.currentIndex != 0 && t.prev() != ';' && !unicode.IsSpace(t.prev()) {
		return
	}

	if string(t.fileRunes[t.currentIndex:t.currentIndex+IMPORT_LENGTH]) != "import" {
		return
	}

	t.advanceChars(IMPORT_LENGTH)
	// in this case import is probably part of a variable name
	if t.current() != '{' && t.current() != '(' && !unicode.IsSpace(t.current()) {
		return
	}

	identifiers := t.tokenizeImportIdentifiers()
	// This will allow some incorrect syntax to be treated as an import, e.g. import './not/a/real/import'
	// But fixing that is low priority for now
	for t.currentIndex < t.end() && !isQuote(t.current()) {
		if t.current() == ';' || t.current() == ')' {
			return
		}
		if t.current() == '/' {
			// Avoids interpreting quotes inside of comments as strings
			t.skipImportComments()
		}
		t.advanceChars()
	}

	importPath, ok := t.readImportString()
	if !ok {
		return
	}

	t.imports[importPath] = identifiers
}

func (t *Tokenizer) tokenizeRequire() {
	if t.currentIndex+REQUIRE_LENGTH > t.end() {
		return
	}

	if t.currentIndex != 0 && t.prev() != ';' && !unicode.IsSpace(t.prev()) {
		return
	}
	// This somewhat breaks doing things in one pass but I suspect the performance drawbacks
	// here aren't significant enough to optimize right away
	if string(t.fileRunes[t.currentIndex:t.currentIndex+REQUIRE_LENGTH]) != "require" {
		return
	}

	t.advanceImportChars(REQUIRE_LENGTH)
	if t.current() != '(' {
		return
	}

	t.advanceImportChars()
	if !isQuote(t.current()) {
		return
	}
	path, ok := t.readImportString()
	if ok {
		t.imports[path] = []string{}
	}
}

func (t *Tokenizer) tokenizeExport() {
	if t.currentIndex+EXPORT_LENGTH > t.end() {
		return
	}

	// the word `export` could be used in a variable name so this protects against that
	if t.currentIndex != 0 && t.prev() != ';' && !unicode.IsSpace(t.prev()) {
		return
	}

	if string(t.fileRunes[t.currentIndex:t.currentIndex+EXPORT_LENGTH]) != "export" {
		return
	}

	t.advanceChars(EXPORT_LENGTH)
	// in this case export is probably part of a variable name
	if t.current() != '{' && !unicode.IsSpace(t.current()) {
		return
	}

	identifiers, isRegularExport := t.tokenizeExportIdentifiers()

	if isRegularExport {
		t.exports = append(t.exports, identifiers...)
		return
	}
	// TODO: This is used multiple times and could probably be abstracted to a `findQuote` function
	// in this case we are dealing with a re-export so we need to find the next string literal
	for current := t.current(); t.currentIndex < t.end() && !isQuote(current); current = t.current() {
		if current == ';' || current == ')' {
			return
		}
		if current == '/' {
			// Avoids interpreting quotes inside of comments as strings
			t.skipImportComments()
		}
		t.advanceChars()
	}

	reExportPath, ok := t.readImportString()
	if !ok {
		return
	}

	t.reExports = append(t.reExports, reExportPath)
	if len(identifiers) == 0 {
		fmt.Printf("WARN: file %q has no reExported identifiers but contained a reExport. This is either a syntax error, or a bug in the tokenizer.\n", t.initPath)
		return
	}

	if t.reExportMap == nil {
		t.reExportMap = make(map[string]string, len(identifiers))
	}
	// populate reExportMap with idents. If an ident is "*"
	// save reExport path in map so that it can be populated in the parser
	for _, ident := range identifiers {
		if ident == "*" {
			t.reExportMap[reExportPath] = "*"
			continue
		}
		t.reExportMap[ident] = reExportPath
	}
}

func (t *Tokenizer) readImportString() (string, bool) {
	var b strings.Builder
	t.advanceChars()
	for t.currentIndex < t.end() && !isQuote(t.current()) {
		b.WriteRune(t.current())
		t.advanceChars()
	}

	if t.currentIndex >= t.end() {
		return "", false
	}

	path := b.String()
	if isRelativePath(path) {
		path = filepath.Join(t.callDir, path)
	}
	return path, true
}

func (t *Tokenizer) skipSingleLineComment() {
	for t.currentIndex < t.end() && t.current() != '\n' {
		t.advanceChars()
	}
}

func (t *Tokenizer) skipMultiLineComment() {
	for t.currentIndex+1 < t.end() && !(t.current() == '*' && t.peek() == '/') {
		t.advanceChars()
	}
	t.advanceChars(2)
}

func (t *Tokenizer) skipWhitespace() {
	for t.currentIndex < t.end() && unicode.IsSpace(t.current()) {
		t.advanceChars()
	}
}

func (t *Tokenizer) skipImportComments() {
	for t.currentIndex < t.end() && t.current() == '/' {
		t.skipComment(t.peek())
		t.skipWhitespace()
	}
}

func (t *Tokenizer) tokenizeImportIdentifiers() []string {
	endChars := []rune{';', '(', '"', '\'', '`'}
	stopChars := []rune{',', '{', '}', '/'}
	var currentIdentifier []rune
	var identifiers []string
	isDefault := true
	for t.currentIndex < t.end() && !slices.Contains(endChars, t.current()) {
		if current := t.current(); slices.Contains(stopChars, current) || unicode.IsSpace(current) {
			ident := string(currentIdentifier)
			switch {
			case ident == "as":
				t.skipNextIdentifier()
			case ident == "from":
				break
			case ident == "type":
				// ignore `type` since it can be used as annotation `import type { FooType } ...`
			case len(ident) > 0:
				if isDefault {
					identifiers = append(identifiers, "default")
				} else {
					identifiers = append(identifiers, ident)
				}
			}

			currentIdentifier = []rune{}

			switch current {
			case '{':
				isDefault = false
			case '}':
				isDefault = true
			case '/':
				if t.peek() == '*' || t.peek() == '/' {
					t.skipImportComments()
					continue
				}
			}
		} else {
			currentIdentifier = append(currentIdentifier, current)
		}

		t.advanceChars()
	}

	return identifiers
}

// returns false when 'from' token is encountered
func (t *Tokenizer) tokenizeExportIdentifiers() ([]string, bool) {
	// TODO: recreating and garbage collecting this on each function call is probably bad for performance
	keywords := []string{"const", "let", "var", "function", "interface", "type"}
	stopChars := []rune{',', '{', '}', '/'}
	endChars := []rune{'=', ';'}

	var currentIdentifier []rune
	var identifiers []string
	// When `default`` is inside of curly braces it means we are re-exporting
	// in this case we need to return false
	haveSeenLeftBrace := false
	// For mapping to work correctly we need to overwrite aliased exports with their alias
	overwriteLastExport := false
	for t.currentIndex < t.end() && !slices.Contains(endChars, t.current()) {
		if current := t.current(); slices.Contains(stopChars, current) || unicode.IsSpace(current) {
			if current == '{' {
				haveSeenLeftBrace = true
			}
			ident := string(currentIdentifier)
			switch ident {
			case "as":
				overwriteLastExport = true
			case "from":
				return identifiers, false
			case "default":
				if !haveSeenLeftBrace {
					return []string{"default"}, true
				}
				// in this case we are in a re-export
				identifiers = addExportIdentifier(identifiers, ident, overwriteLastExport)
				overwriteLastExport = false
			case "interface":
				// interfaces are of form export interface ident { //... }
				endChars = append(endChars, '{')
			default:
				if !slices.Contains(keywords, ident) && len(ident) > 0 {
					identifiers = addExportIdentifier(identifiers, ident, overwriteLastExport)
					overwriteLastExport = false
				}
			}

			currentIdentifier = []rune{}
			if current == '/' && (t.peek() == '/' || t.peek() == '*') {
				t.skipImportComments()
				continue
			}
		} else {
			currentIdentifier = append(currentIdentifier, current)
		}

		t.advanceChars()
	}

	return identifiers, true
}

// Assumes correct syntax. This would fail for the case `as from "./path";`
func (t *Tokenizer) skipNextIdentifier() {
	endChars := []rune{';', '(', ',', '{', '}'}
	t.skipWhitespace()
	for current := t.current(); t.currentIndex < t.end() && !slices.Contains(endChars, current) && !unicode.IsSpace(current); current = t.current() {
		if current == '/' && (t.peek() == '*' || t.peek() == '/') {
			t.skipImportComments()
		}
		t.advanceChars()
	}
}

func (t *Tokenizer) peek() rune {
	// TODO: Refactor this to be less sloppy
	if t.currentIndex+1 >= t.end() {
		return 0
	}
	return t.fileRunes[t.currentIndex+1]
}

func (t *Tokenizer) current() rune {
	if t.currentIndex >= t.end() {
		return 0
	}
	return t.fileRunes[t.currentIndex]
}

// This is useful for import to determine if it is part of a variable name
// It's not strictly necessary because we could use the peek() to accomplish
// this as well but for now it's making my life easier.
func (t *Tokenizer) prev() rune {
	if t.currentIndex < 1 {
		return 0
	}
	return t.fileRunes[t.currentIndex-1]
}

func (t *Tokenizer) end() int {
	return len(t.fileRunes)
}

// Advances chars and skips whitespace and comments
func (t *Tokenizer) advanceImportChars(args ...int) {
	step := 1
	if len(args) > 0 && args[0] > 0 {
		step = args[0]
	}
	t.advanceChars(step)
	t.skipWhitespace()
	t.skipImportComments()
}

// Supports optional int to advance by more than one. Subsequent ints are ignored.
func (t *Tokenizer) advanceChars(args ...int) {
	step := 1
	if len(args) > 0 && args[0] > 0 {
		step = args[0]
	}
	t.currentIndex += step
}

func isQuote(c rune) bool {
	return c == '\'' || c == '"' || c == '`'
}

func isRelativePath(path string) bool {
	return strings.HasPrefix(path, ".")
}

// Performance note (because I'm bad at Go): Slices are passed by value but a slice is
// simply a header that gives information about an underlying array (e.g. pointer to array
// plus start and stop indices). So reassigning a slice is not nearly as expensive as
// copying and reassigning an array in languages like JavaScript
func addExportIdentifier(identifiers []string, ident string, overwrite bool) []string {
	if overwrite {
		identifiers[len(identifiers)-1] = ident
	} else {
		identifiers = append(identifiers, ident)
	}
	return identifiers
}
