package tokenizer

import (
	"slices"
	"strings"
	"unicode"
)

// Export keywords. See: https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Statements/export
// 	"var",
// 	"let",
// 	"const",
// 	"function",
// 	"class",
// 	"function*",
// 	"type",
// 	"interface",

// There are more characters in javascript that could signify
// we're at the end of an identifier but not all are relevant
// for import tokenizing.
var identifier_ends = []rune{
	'{',
	'}',
	';',
	',',
	'(',
	'/',
	// Note: these should only show up in exports
	'[',
	']',
	':',
}

func isIdentifierEnd(char rune) bool {
	return unicode.IsSpace(char) || slices.Contains(identifier_ends, char)
}

func isQuote(char rune) bool {
	return char == '\'' || char == '"' || char == '`'
}

func isRelativePath(path string) bool {
	return strings.HasPrefix(path, ".")
}
