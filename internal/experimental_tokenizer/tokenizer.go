package experimental_tokenizer

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"unicode"

	"github.com/stilt0n/dependor/internal/tokenizer"
)

type Tokenizer struct {
	// points to index of current `char`
	currentIndex int
	// points to index of next character to read
	readIndex   int
	char        rune
	fileRunes   []rune
	imports     map[string][]string
	reExports   []string
	reExportMap map[string]string
	exports     []string
	callDir     string
	initPath    string
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
		currentIndex: -1,
		readIndex:    0,
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
			// comments could contain keywords in them but should not be parsed as imports / exports
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

// Export cases: https://developer.mozilla.org/en-US/docs/web/javascript/reference/statements/export
func (t *Tokenizer) readExport() {
	var identifiers []string
	isReExport := false
	haveSeenLeftBracket := false
	// '}' is also sort of an an endChar. e.g. export { foo, bar } w/o semi-colon
	// but it's ambiguous because for re-exports we'd need to continue. I'm not sure
	// there's any way to handle this case besides looking ahead afterwards
	endedOnBracket := false
	// overwrite exported identifiers with their aliases so that they are correctly mapped to importing files
	overwriteLastIdentifier := false
	endChars := []rune{';', '(', '='}

Loop:
	for t.char != 0 {
		switch {
		// this needs to come before isIdentifierEnd check because some of these chars are shared
		case slices.Contains(endChars, t.char):
			break Loop
		case t.char == '{':
			haveSeenLeftBracket = true
			t.readChar()
		case t.char == '/':
			t.skipComment(false)
		case t.char == '}':
			endedOnBracket = true
			break Loop
		case isIdentifierEnd(t.char):
			t.readChar()
		case isQuote(t.char):
			panic(fmt.Sprintf("Encountered a quote in an export statement that was not preceded by the `from` keyword in %q. This may be a syntax error or it could be a quoted export alias. Dependor does not currently support quoted import aliases.", t.initPath))
		default:
			ident := t.readIdentifier()
			// I don't think this case can happen, but if it does, this will avoid an infinite loop
			if len(ident) == 0 && t.char != 0 {
				panic(fmt.Sprintf("Unexpected length 0 identifier not located at the end of the file in %q. This situation was expected to be impossible.", t.initPath))
			}
			switch ident {
			case "as":
				overwriteLastIdentifier = true
			case "from":
				isReExport = true
				break Loop
			case "interface":
				endChars = append(endChars, '{')
			case "const", "let", "var", "function", "function*", "class", "type":
				continue
			default:
				if ident == "default" && !haveSeenLeftBracket {
					identifiers = append(identifiers, ident)
					break Loop
				}
				if overwriteLastIdentifier {
					identifiers[len(identifiers)-1] = ident
					overwriteLastIdentifier = false
					continue
				}
				identifiers = append(identifiers, ident)
			}
		}
	}
	if endedOnBracket {
		t.readChar()
		t.skipAllFiller()
		if t.char == 'f' {
			ident := t.readIdentifier()
			if ident == "from" {
				isReExport = true
			}
		}
	}

	if !isReExport {
		t.exports = append(t.exports, identifiers...)
		return
	}

	if t.reExportMap == nil {
		t.reExportMap = make(map[string]string, 0)
	}

	t.skipAllFiller()
	if !isQuote(t.char) {
		panic(fmt.Sprintf("Unexpected non-string token following the keyword `from` in %q. This is likely due to a syntax error.", t.initPath))
	}
	reExportPath := t.readPathString()

	t.reExports = append(t.reExports, reExportPath)
	if len(identifiers) == 0 {
		panic(fmt.Sprintf("Unexpected re-export with zero identifiers in %q. This is likely a syntax error.", t.initPath))
	}

	// populate reExportMap with idents. If an ident is "*"
	// save reExport path in map so that it can be populated
	// later inside of the parser. For aliased *'s (e.g. export * as namespace from './file')
	// then mapping the alias to the file path should be sufficient
	for _, ident := range identifiers {
		if ident == "*" {
			t.reExportMap[reExportPath] = "*"
			continue
		}
		t.reExportMap[ident] = reExportPath
	}
}

// Note: this can end up tokenizing invalid javascript e.g. import +not+a+valid+var from 'path';
// We try to throw errors on these cases when possible but checking a valid javascript variable
// is pretty complex because the spec allows many unicode characters including emojis. Instead
// we just assume any non-whitespace character not in identifier_ends is a valid identifier char
func (t *Tokenizer) readImport() {
	var identifiers []string
	skipNextIdentifier := false
	// used to determine if import is a default import
	haveSeenLeftBracket := false
	for t.char != 0 {
		switch {
		case t.char == '/':
			t.skipComment(false)
		case isIdentifierEnd(t.char):
			if t.char == '{' {
				haveSeenLeftBracket = true
			}
			t.readChar()
		case t.char == ')':
			// import() can import using a variable rather than a string I think
			return
		case isQuote(t.char):
			importPath := t.readPathString()
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
			// Some typescript setups annotate imports of types as `import type { ... } ...`
			case "from", "type":
				continue
			default:
				if skipNextIdentifier {
					skipNextIdentifier = false
					continue
				}
				// non-bracketed imports are either default imports or namespace imports
				if !haveSeenLeftBracket && ident != "*" {
					ident = "default"
				}
				identifiers = append(identifiers, ident)
			}
		}
	}
	panic(fmt.Sprintf("Encountered non-terminating import statement in %q. This is likely a syntax error.", t.initPath))
}

// skips to first non-whitespace non-comment character
func (t *Tokenizer) skipAllFiller() {
Loop:
	for t.char != 0 {
		switch {
		case t.char == '/':
			t.skipComment(false)
		case unicode.IsSpace(t.char):
			t.skipWhitespace()
		default:
			break Loop
		}
	}
}

func (t *Tokenizer) readRequire() {
	for t.char != 0 {
		switch {
		case t.char == ')':
			return
		case t.char == '/':
			t.skipComment(false)
		case isQuote(t.char):
			requirePath := t.readPathString()
			t.imports[requirePath] = []string{}
			return
		default:
			t.readChar()
		}
	}
	panic(fmt.Sprintf("Encountered a non-terminating require statement in %q. This is likely a syntax error.", t.initPath))
}

// skips to first non-whitespace character
func (t *Tokenizer) skipWhitespace() {
	for unicode.IsSpace(t.char) {
		t.readChar()
	}
}

// we try to panic on syntax errors when it's reasonable to do so rather than give incorrect output:
// when slashes show up in import / export statements they should be 1) part of a string or 2) part of a comment.
// when slashes show up outside of import / export statements then they are valid JavaScript and we should not panic
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

// skips to first character that is not part of a single line comment
func (t *Tokenizer) skipSingleLineComment() {
	for t.char != 0 && t.char != '\n' && t.char != '\r' {
		t.readChar()
	}
	if t.char != 0 {
		t.readChar()
	}
}

// Skips to first character that is not part of a multilinie comment
func (t *Tokenizer) skipMultiLineComment() {
	t.readChar()
	for t.char != 0 {
		if t.char == '*' && t.peek() == '/' {
			break
		}
		t.readChar()
	}
	t.readChar()
	t.readChar()
}

func (t *Tokenizer) readIdentifier() string {
	start := t.currentIndex
	for t.char != 0 && !isIdentifierEnd(t.char) && !isQuote(t.char) {
		t.readChar()
	}
	return string(t.fileRunes[start:t.currentIndex])
}

func (t *Tokenizer) readPathString() string {
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

	pathString := string(t.fileRunes[start:t.currentIndex])
	if isRelativePath(pathString) {
		pathString = filepath.Join(t.callDir, pathString)
	}
	return pathString
}

func (t *Tokenizer) readChar() {
	if t.readIndex >= t.end() {
		t.char = 0
		return
	}
	t.char = t.fileRunes[t.readIndex]
	t.readIndex++
	t.currentIndex++
}

func (t *Tokenizer) end() int {
	return len(t.fileRunes)
}

func (t *Tokenizer) peek() rune {
	if t.readIndex == len(t.fileRunes) {
		return 0
	}
	return t.fileRunes[t.readIndex]
}
