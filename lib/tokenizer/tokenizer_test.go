package tokenizer

import (
	"testing"
)

func TestTerminates(t *testing.T) {
	tk := New(`const foo = 5;`)
	output := tk.TokenizeImports()
	if len(output) != 0 {
		t.Fatalf("Should not be any import tokens")
	}
}

func TestSimpleRequire(t *testing.T) {
	tokenizer := New(`const foo = require("./foo");`)
	output := tokenizer.TokenizeImports()
	if len(output) != 1 {
		t.Fatalf("Expected output to be length 1. Got %d", len(output))
	}
	if output[0] != "./foo" {
		t.Fatalf(`Expected "./foo". Got %s`, output[0])
	}
}

func TestImportComments(t *testing.T) {
	tokenizer := New(`const igloo = require/* rude */  /* ugh*/( /* why */"./igloo");`)
	output := tokenizer.TokenizeImports()
	if len(output) != 1 {
		t.Fatalf("Expected output to be length 1. Got %d", len(output))
	}
	if output[0] != "./igloo" {
		t.Fatalf(`Expected "./igloo". Got %s`, output[0])
	}
}

func TestSimpleImport(t *testing.T) {
	tokenizer := New(`import foo from "./foo";`)
	output := tokenizer.TokenizeImports()
	if len(output) != 1 {
		t.Fatalf("Expected output to be length 1. Got %d", len(output))
	}
	if output[0] != "./foo" {
		t.Fatalf(`Expected "./foo". Got %s`, output[0])
	}
}

func TestDynamicImport(t *testing.T) {
	tokenizer := New(`const foo = await import("./foo");`)
	output := tokenizer.TokenizeImports()
	if len(output) != 1 {
		t.Fatalf("Expected output to be length 1. Got %d", len(output))
	}
	if output[0] != "./foo" {
		t.Fatalf(`Expected "./foo". Got %s`, output[0])
	}
}

func TestInvalidImport(t *testing.T) {
	tokenizer := New(`import hello there`)
	output := tokenizer.TokenizeImports()
	if len(output) != 0 {
		t.Fatalf("Expected no imports to be output. Got %s", output[0])
	}
}

func TestTokenizeFile(t *testing.T) {
	tokenizer, err := NewTokenizerFromFile("./test.js")
	if err != nil {
		t.Fatalf("Expected successful file read. Got error: %s", err)
	}
	output := tokenizer.TokenizeImports()
	expected := []string{
		"fs",
		"foo",
		"../components/bar",
		"../noSemicolon/alphabet",
		"./dir/path/file",
		".",
		"./example",
		"polite",
		"~/path",
		"../lib",
		"./a/long/path/that/might/fit/better/on/mutliple/lines/i/guess",
		"./space/bar.json",
		"tricky",
	}
	for i, imp := range output {
		if imp != expected[i] {
			t.Errorf("Error in example %d.\n  Got: %s\n  Expected: %s", i, imp, expected[i])
		}
	}
}
