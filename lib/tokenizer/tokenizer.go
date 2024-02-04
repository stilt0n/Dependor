/*
Regex seems to mostly work for this but is a little challenging when it comes to handling comments
especially since Go does not support lookbehind regex (for performance reasons, apparently).
I didn't think it would be crazy difficult to write an import lexer so that's what this is.
*/
package tokenizer

import (
	"os"
	"strings"
	"unicode"
)

const (
	REQUIRE_LENGTH = 7
	IMPORT_LENGTH  = 6
)

type Tokenizer struct {
	// We want a nextIndex pointer so we can read comments
	currentIndex, nextIndex int
	fileRunes               []rune
	imports                 []string
}

func NewTokenizerFromFile(filepath string) (*Tokenizer, error) {
	file, err := os.ReadFile(filepath)
	if err != nil {
		return &Tokenizer{}, err
	}
	fileString := string(file)
	return New(fileString), nil
}

func New(fileString string) *Tokenizer {
	t := Tokenizer{
		currentIndex: 0,
		nextIndex:    1,
		fileRunes:    []rune(fileString),
		imports:      []string{},
	}
	return &t
}

// Something to keep in mind: this will at least miss *some* edge cases.
// e.g. nested requires `require(require('./pathToRealImport'))`
// Supporting this case seems challenging and I don't think it's currently worth my effort
// Reads all import paths from a file in one pass and returns them in an array
func (t *Tokenizer) TokenizeImports() []string {
	if len(t.fileRunes) < 1 {
		return t.imports
	}

	for t.currentIndex < t.end() {
		char := t.current()
		switch char {
		case 'i':
			t.tokenizeImport()
		case 'r':
			t.tokenizeRequire()
		case '/':
			t.skipComment(t.peek())
		}
		t.advanceChars()
	}
	return t.imports
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

	t.readImportString()
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
	t.readImportString()
}

func (t *Tokenizer) readImportString() {
	var b strings.Builder
	t.advanceChars()
	for t.currentIndex < t.end() && !isQuote(t.current()) {
		b.WriteRune(t.current())
		t.advanceChars()
	}

	if t.currentIndex >= t.end() {
		return
	}

	t.imports = append(t.imports, b.String())
}

func (t *Tokenizer) skipSingleLineComment() {
	for t.currentIndex < t.end() && t.current() != '\n' {
		t.advanceChars()
	}
}

func (t *Tokenizer) skipMultiLineComment() {
	for t.nextIndex < t.end() && !(t.current() == '*' && t.peek() == '/') {
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
		if t.peek() == '*' {
			t.skipMultiLineComment()
			t.skipWhitespace()
		}
	}
}

func (t *Tokenizer) peek() rune {
	// TODO: Refactor this to be less sloppy
	if t.nextIndex >= t.end() {
		return 0
	}
	return t.fileRunes[t.nextIndex]
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
	t.nextIndex += step
}

func isQuote(c rune) bool {
	return c == '\'' || c == '"' || c == '`'
}
