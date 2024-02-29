package experimental_tokenizer

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"unicode"

	"github.com/stilt0n/dependor/internal/tokenizer"
)

// There are more characters in javascript that could signify
// we're at the end of an identifier but not all are relevant
// for import tokenizing.
var identifier_ends = []rune{
	'{',
	'}',
	';',
	',',
	'(',
}

type Tokenizer struct {
	currentIndex int
	char         rune
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
	t.readChar()
	return &t
}

func (t *Tokenizer) Tokenize() tokenizer.FileToken {
	if len(t.fileRunes) < 1 {
		return tokenizer.FileToken{}
	}

	for t.char != 0 {
		if t.char == 'i' || t.char == 'r' || t.char == 'e' {
			switch token := t.readIdentifier(); token {
			case "import":
				t.readImport()
			case "require":
				t.readRequire()
			case "export":
				t.readExport()
			}
		} else if t.char == '/' {
			t.skipComment(true)
		} else {
			// avoids reading char if a previous read got us to EOF
			t.readChar()
		}
	}

	return tokenizer.FileToken{
		FilePath:    t.initPath,
		Imports:     t.imports,
		ReExports:   t.reExports,
		Exports:     t.exports,
		ReExportMap: t.reExportMap,
	}
}

// Note: this can end up tokenizing invalid javascript e.g. import +not+a+valid+var from 'path';
// I try to throw errors on these cases when possible but checking a valid javascript variable
// is pretty complex because the spec allows many unicode characters including emojis. Instead
// we just assume any non-whitespace character not in identifier_ends is a valid identifier char
func (t *Tokenizer) readImport() {
	var identifiers []string
	skipNextIdentifier := false
	for t.char != 0 {
		switch {
		case isIdentifierEnd(t.char):
			t.readChar()
		case t.char == '/':
			t.skipComment(false)
		case isQuote(t.char):
			importPath := t.readString()
			t.imports[importPath] = append(t.imports[importPath], identifiers...)
			return
		default:
			ident := t.readIdentifier()
			// I don't think this case can happen, but if it does, this will avoid an infinite loop
			if len(ident) == 0 && t.char != 0 {
				panic(fmt.Sprintf("Unexpected length 0 identifier not located at the end of the file in %q. This situation was expected to be impossible.", t.initPath))
			}

			switch ident {
			case "as":
				skipNextIdentifier = true
			case "from":
				continue
			default:
				if skipNextIdentifier {
					skipNextIdentifier = false
					continue
				}
				identifiers = append(identifiers, ident)
			}
		}
	}
	panic(fmt.Sprintf("Encountered non-terminating import statement in %q. This is likely a syntax error.", t.initPath))
}

func (t *Tokenizer) readRequire() {
	for t.char != 0 {
		switch {
		case t.char == ')':
			return
		case t.char == '/':
			t.skipComment(false)
		case isQuote(t.char):
			requirePath := t.readString()
			t.imports[requirePath] = []string{}
			return
		default:
			t.readChar()
		}
	}
	panic(fmt.Sprintf("Encountered a non-terminating require statement in %q. This is likely a syntax error.", t.initPath))
}

func (t *Tokenizer) skipWhitespace() {
	for unicode.IsSpace(t.char) {
		t.readChar()
	}
}

func (t *Tokenizer) skipComment(standaloneSlashIsSafe bool) {
	t.readChar()
	switch t.char {
	case '/':
		t.skipSingleLineComment()
	case '*':
		t.skipMultiLineComment()
	default:
		// Unless this function is called at the top level, standalone slashes are invalid syntax.
		if !standaloneSlashIsSafe {
			panic(fmt.Sprintf("Error: tokenizer came across an unexpected '/' character in an import or export statement that is not part of an import string or comment. This occurred in %q and is likely due to a syntax error.\n", t.initPath))
		}
	}
}

func (t *Tokenizer) skipSingleLineComment() {
	for t.char != 0 && t.char != '\n' && t.char != '\r' {
		t.readChar()
	}
}

func (t *Tokenizer) skipMultiLineComment() {
	t.readChar()
	for t.char != 0 {
		if t.char == '*' && t.peek() == '/' {
			break
		}
		t.readChar()
	}
}

func (t *Tokenizer) readIdentifier() string {
	start := t.currentIndex
	for t.char != 0 && !isIdentifierEnd(t.char) {
		t.readChar()
	}
	return string(t.fileRunes[start:t.currentIndex])
}

func (t *Tokenizer) readString() string {
	// should be starting on a quote so we need to advance to first nonquote
	t.readChar()
	start := t.currentIndex
	for t.currentIndex < t.end() {
		if isQuote(t.char) {
			break
		}
		t.readChar()
	}

	if t.currentIndex == t.end() {
		// Panic here since syntax errors like this could silently cause the
		// resultant import graph to be incorrect.
		panic(fmt.Sprintf("Error: tokenizer came across a non-terminating string in %q. This is likely a syntax error.\n", t.initPath))
	}

	return string(t.fileRunes[start:t.currentIndex])
}

func (t *Tokenizer) readChar() {
	if t.currentIndex >= t.end() {
		t.char = 0
	} else {
		t.char = t.fileRunes[t.currentIndex]
	}
	t.currentIndex++
}

func (t *Tokenizer) end() int {
	return len(t.fileRunes)
}

func (t *Tokenizer) peek() rune {
	if t.currentIndex+1 == len(t.fileRunes) {
		return 0
	}
	return t.fileRunes[t.currentIndex+1]
}

func isIdentifierEnd(char rune) bool {
	return unicode.IsSpace(char) || slices.Contains(identifier_ends, char)
}

func isQuote(char rune) bool {
	return char == '\'' || char == '"' || char == '`'
}
