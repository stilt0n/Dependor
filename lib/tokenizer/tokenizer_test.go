package tokenizer

import (
	"slices"
	"testing"
)

func TestTerminates(t *testing.T) {
	tk := New(`const foo = 5;`, "./testfiles")
	result := tk.TokenizeImports()
	output := result.ImportStrings()
	if len(output) != 0 {
		t.Fatalf("Should not be any import tokens")
	}
}

func TestSimpleRequire(t *testing.T) {
	tokenizer := New(`const foo = require("./foo");`, ".")
	result := tokenizer.TokenizeImports()
	output := result.ImportStrings()
	if len(output) != 1 {
		t.Fatalf("Expected output to be length 1. Got %d", len(output))
	}
	if output[0] != "foo" {
		t.Fatalf(`Expected "foo". Got %s`, output[0])
	}
}

func TestImportComments(t *testing.T) {
	tokenizer := New(`const igloo = require/* rude */  /* ugh*/( /* why */"./igloo");`, ".")
	result := tokenizer.TokenizeImports()
	output := result.ImportStrings()
	if len(output) != 1 {
		t.Fatalf("Expected output to be length 1. Got %d", len(output))
	}
	if output[0] != "igloo" {
		t.Fatalf(`Expected "igloo". Got %s`, output[0])
	}
}

func TestSimpleImport(t *testing.T) {
	tokenizer := New(`import foo from "./foo";`, ".")
	result := tokenizer.TokenizeImports()
	output := result.ImportStrings()
	if len(output) != 1 {
		t.Fatalf("Expected output to be length 1. Got %d", len(output))
	}
	if output[0] != "foo" {
		t.Fatalf(`Expected "foo". Got %s`, output[0])
	}
}

func TestDynamicImport(t *testing.T) {
	tokenizer := New(`const foo = await import("./foo");`, ".")
	result := tokenizer.TokenizeImports()
	output := result.ImportStrings()
	if len(output) != 1 {
		t.Fatalf("Expected output to be length 1. Got %d", len(output))
	}
	if output[0] != "foo" {
		t.Fatalf(`Expected "foo". Got %s`, output[0])
	}
}

func TestInvalidImport(t *testing.T) {
	tokenizer := New(`import hello there`, ".")
	result := tokenizer.TokenizeImports()
	output := result.ImportStrings()
	if len(output) != 0 {
		t.Fatalf("Expected no imports to be output. Got %s", output[0])
	}
}

func TestTokenizeFile(t *testing.T) {
	tokenizer, err := NewTokenizerFromFile("./testfiles/nested/test.js")
	if err != nil {
		t.Fatalf("Expected successful file read. Got error: %s", err)
	}
	result := tokenizer.TokenizeImports()
	output := result.ImportStrings()
	expected := []string{
		"fs",
		"foo",
		"testfiles/components/bar",
		"testfiles/noSemicolon/alphabet",
		"testfiles/nested/dir/path/file",
		"testfiles/nested",
		"testfiles/nested/example",
		"polite",
		"~/path",
		"testfiles/lib",
		"testfiles/nested/a/long/path/that/might/fit/better/on/mutliple/lines/i/guess",
		"testfiles/nested/space/bar.json",
		"tricky",
	}
	slices.Sort(expected)
	slices.Sort(output)
	for i, imp := range output {
		if imp != expected[i] {
			t.Errorf("Error in example %d.\n  Got: %s\n  Expected: %s", i, imp, expected[i])
		}
	}
}

func TestTokenizeIdentifiers(t *testing.T) {
	expected := map[string][]string{
		"testfiles/foo":        {"default", "a", "b", "c"},
		"testfiles/nested/bar": {"item"},
		"testfiles/nested/baz": {"ident", "bar"},
		"just-the-path":        {},
	}
	tokenizer, err := NewTokenizerFromFile("./testfiles/nested/test_idents.js")
	if err != nil {
		t.Fatalf("Expected successful file read. Got error: %s", err)
	}
	tokenizedFile := tokenizer.TokenizeImports()

	if len(tokenizedFile.Imports) != len(expected) {
		t.Fatalf("Number of imports (%d) does not match expected number (%d)", len(tokenizedFile.Imports), len(expected))
	}

	for pth, idents := range tokenizedFile.Imports {
		expectedIdents, ok := expected[pth]
		if !ok {
			t.Fatalf("Received path %q which is not in imported paths", pth)
		}

		if len(idents) != len(expectedIdents) {
			t.Fatalf("Wrong number of identifiers for path %q. Expected %d received %d", pth, len(expectedIdents), len(idents))
		}

		for i, ident := range idents {
			if ident != expectedIdents[i] {
				t.Errorf("Expected %q for identifier at index %d but received %q", expectedIdents[i], i, ident)
			}
		}
	}
}

func TestTokenizeExports(t *testing.T) {
	tokenizer, err := NewTokenizerFromFile("./testfiles/nested/test2.js")
	if err != nil {
		t.Fatalf("Expected successful file read. Got error: %s", err)
	}

	expectedExports := []string{
		"five",
		"foo",
		"bar",
		"baz",
		"default",
	}

	expectedImportIdents := []string{"default", "example"}

	tokenizedFile := tokenizer.TokenizeImports()

	if len(tokenizedFile.Imports) != 1 {
		t.Fatalf("Wrong length for imports. Expected 1 received %d", len(tokenizedFile.Imports))
	}

	exampleIdents, ok := tokenizedFile.Imports["example"]

	if !ok {
		t.Fatal("Expected \"example\" to be in import paths")
	}

	for i, ident := range exampleIdents {
		if ident != expectedImportIdents[i] {
			t.Errorf("Wrong import for index %d. Expected %q received %q", i, expectedImportIdents[i], ident)
		}
	}

	if len(tokenizedFile.Exports) != len(expectedExports) {
		t.Fatalf("Expected exports length to be %d but received length %d", len(expectedExports), len(tokenizedFile.Exports))
	}

	for i, ident := range tokenizedFile.Exports {
		if ident != expectedExports[i] {
			t.Errorf("Wrong export for index %d. Expected %q received %q", i, expectedExports[i], ident)
		}
	}
}

func TestReExports(t *testing.T) {
	expectedReExports := []string{
		"./test",
		"./test2",
	}

	tokenizer, err := NewTokenizerFromFile("./testfiles/nested/index.js")
	if err != nil {
		t.Fatalf("Expected successful file read. Got error: %s", err)
	}

	tokenizedFile := tokenizer.TokenizeImports()

	if len(tokenizedFile.ReExports) != len(expectedReExports) {
		t.Fatalf("Expected %d re-exports but received %d", len(expectedReExports), len(tokenizedFile.ReExports))
	}

	for i, rex := range tokenizedFile.ReExports {
		if rex != expectedReExports[i] {
			t.Errorf("Expected re-export at index %d to be %q but received %q", i, expectedReExports[i], rex)
		}
	}
}
