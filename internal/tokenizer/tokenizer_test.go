package tokenizer

import (
	"slices"
	"testing"
)

func TestTerminates(t *testing.T) {
	tk := New(`const foo = 5;`, "./testfiles")
	tokenizedFile := tk.Tokenize()
	output := getImportStrings(tokenizedFile)
	if len(output) != 0 {
		t.Fatalf("Should not be any import tokens")
	}
}

func TestTerminatesOnBracedExportWithoutSemicolon(t *testing.T) {
	tk := New(`export { foo, bar }`, ".")
	tokenizedFile := tk.Tokenize()
	expected := []string{"foo", "bar"}
	testArray(t, tokenizedFile.Exports, expected)
}

func TestSimpleRequire(t *testing.T) {
	tokenizer := New(`const foo = require("./foo");`, ".")
	tokenizedFile := tokenizer.Tokenize()
	output := getImportStrings(tokenizedFile)
	if len(output) != 1 {
		t.Fatalf("Expected output to be length 1. Got %d", len(output))
	}
	if output[0] != "foo" {
		t.Fatalf(`Expected "foo". Got %s`, output[0])
	}
}

func TestImportComments(t *testing.T) {
	tokenizer := New(`const igloo = require/* rude */  /* ugh*/( /* why */"./igloo");`, ".")
	tokenizedFile := tokenizer.Tokenize()
	output := getImportStrings(tokenizedFile)
	if len(output) != 1 {
		t.Fatalf("Expected output to be length 1. Got %d", len(output))
	}
	if output[0] != "igloo" {
		t.Fatalf(`Expected "igloo". Got %s`, output[0])
	}
}

func TestSimpleImport(t *testing.T) {
	tokenizer := New(`import foo from "./foo";`, ".")
	tokenizedFile := tokenizer.Tokenize()
	output := getImportStrings(tokenizedFile)
	if len(output) != 1 {
		t.Fatalf("Expected output to be length 1. Got %d", len(output))
	}
	if output[0] != "foo" {
		t.Fatalf(`Expected "foo". Got %s`, output[0])
	}
}

func TestDynamicImport(t *testing.T) {
	tokenizer := New(`const foo = await import("./foo"); "bar";`, ".")
	tokenizedFile := tokenizer.Tokenize()
	output := getImportStrings(tokenizedFile)
	if len(output) != 1 {
		t.Fatalf("Expected output to be length 1. Got %d", len(output))
	}
	if output[0] != "foo" {
		t.Fatalf(`Expected "foo". Got %s`, output[0])
	}
}

func TestNonTerminatingImport(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("Expected panic on non-terminating import\n")
		}
	}()
	tokenizer := New(`import hello there`, ".")
	tokenizer.Tokenize()
}

func TestInterfaceExport(t *testing.T) {
	tokenizer := New(`export interface EdgeCase extends Pick<Foo, 'bar' | 'baz'> {};`, "*")
	tokenizedFile := tokenizer.Tokenize()
	testArray(t, tokenizedFile.Exports, []string{"EdgeCase"})
}

func TestTokenizeFile(t *testing.T) {
	tokenizer, err := NewTokenizerFromFile("./testfiles/nested/test.js")
	if err != nil {
		t.Fatalf("Expected successful file read. Got error: %s", err)
	}
	tokenizedFile := tokenizer.Tokenize()
	output := getImportStrings(tokenizedFile)
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
	if len(output) != len(expected) {
		t.Errorf("Expected output to have %d imports but received %d\n", len(expected), len(output))
	}
	slices.Sort(expected)
	slices.Sort(output)
	for i, imp := range output {
		if imp != expected[i] {
			t.Errorf("Error in example %d.\n  Got: %s\n  Expected: %s", i, imp, expected[i])
		}
	}
	t.Log(output)
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
	tokenizedFile := tokenizer.Tokenize()

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

func TestImportTypes(t *testing.T) {
	tokenizer, err := NewTokenizerFromFile("./testfiles/nested/test2.ts")
	if err != nil {
		t.Fatalf("Expected successful file read. Got error: %s", err)
	}

	expected := map[string][]string{
		"example":  {"default", "example"},
		"@Foo/foo": {"FooType", "Foo"},
	}

	tokenizedFile := tokenizer.Tokenize()

	testEdgeList(t, tokenizedFile.Imports, expected)
	t.Logf("%+v\n", tokenizedFile.Imports)
}

func TestTokenizeExports(t *testing.T) {
	tokenizer, err := NewTokenizerFromFile("./testfiles/nested/test2.ts")
	if err != nil {
		t.Fatalf("Expected successful file read. Got error: %s", err)
	}

	expectedExports := []string{
		"x",
		"fun",
		"funner",
		"five",
		"pressF",
		"bar",
		"baz",
		"Noop",
		"IStuff",
		"default",
	}

	tokenizedFile := tokenizer.Tokenize()

	if len(tokenizedFile.Exports) != len(expectedExports) {
		t.Logf("%+v\n", tokenizedFile.Exports)
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
		"testfiles/nested/test",
		"testfiles/nested/test2",
	}

	expectedReExportMap := map[string]string{
		"func":                   "testfiles/nested/test",
		"testfiles/nested/test2": "*",
	}

	tokenizer, err := NewTokenizerFromFile("./testfiles/nested/index.js")
	if err != nil {
		t.Fatalf("Expected successful file read. Got error: %s", err)
	}

	tokenizedFile := tokenizer.Tokenize()

	if len(tokenizedFile.ReExports) != len(expectedReExports) {
		t.Logf("%+v\n", tokenizedFile.ReExports)
		t.Fatalf("Expected %d re-exports but received %d", len(expectedReExports), len(tokenizedFile.ReExports))
	}

	for i, rex := range tokenizedFile.ReExports {
		if rex != expectedReExports[i] {
			t.Errorf("Expected re-export at index %d to be %q but received %q", i, expectedReExports[i], rex)
		}
	}

	if len(tokenizedFile.ReExportMap) != len(expectedReExportMap) {
		t.Log(tokenizedFile.ReExportMap)
		t.Fatalf("Expected %d entries in the re-export map but received %d", len(expectedReExportMap), len(tokenizedFile.ReExportMap))
	}
	for k, v := range tokenizedFile.ReExportMap {
		expectedValue, keyExists := expectedReExportMap[k]
		if !keyExists {
			t.Fatalf("Received unexpected key %q from reExportMap", k)
		}
		if expectedValue != v {
			t.Errorf("Received wrong value for key %q. Expected %q received %q", k, expectedValue, v)
		}
	}
}

func TestMdnImports(t *testing.T) {
	expected := map[string][]string{
		"module-name0": {"default"},
		"module-name1": {"*"},
		"module-name2": {"export1"},
		"module-name3": {"export1"},
		"module-name4": {"default"},
		"module-name5": {"export1", "export2"},
		"module-name6": {"export1", "export2"},
		"module-name7": {"default", "export1"},
		"module-name8": {"default", "*"},
		"module-name9": {},
	}

	tokenizer, err := NewTokenizerFromFile("./testfiles/mdn-import-examples.js")
	if err != nil {
		t.Fatalf("Expected successful file read. Got error: %s", err)
	}

	tokenizedFile := tokenizer.Tokenize()

	testEdgeList(t, tokenizedFile.Imports, expected)
}

func TestMdnExports(t *testing.T) {
	expectedExports := []string{
		"functionName",
		"ClassName",
		"generatorFunctionName",
		"name1",
		"bar",
		"name1",
		"name2",
		"name1",
		"nameN",
		"name1",
		"name2",
		"nameN",
		"default",
		"default",
		"default",
		"default",
		"default",
		"default",
		"default",
		"default",
	}

	tokenizer, err := NewTokenizerFromFile("./testfiles/mdn-export-examples.js")
	if err != nil {
		t.Fatalf("Expected successful file read. Got error: %s", err)
	}

	tokenizedFile := tokenizer.Tokenize()
	testArray(t, tokenizedFile.Exports, expectedExports)
}

func testEdgeList(t *testing.T, edgeList, expected map[string][]string) {
	if len(edgeList) != len(expected) {
		t.Errorf("Expected edge list to have length %d but receive %d", len(expected), len(edgeList))
		return
	}

	for node, edges := range edgeList {
		expectedEdges, ok := expected[node]
		if !ok {
			t.Errorf("Unexpected node %q in edge list", node)
			continue
		}

		if len(expectedEdges) != len(edges) {
			t.Errorf("Expected node %q to have %d edges but received %d", node, len(expectedEdges), len(edges))
			continue
		}

		for i, e := range edges {
			if e != expectedEdges[i] {
				t.Errorf("Expected edge at index %d to be %q but received %q instead.", i, expectedEdges[i], e)
			}
		}
	}
}

func testArray(t *testing.T, arr, expected []string) {
	if len(arr) != len(expected) {
		t.Logf("%+v\n", arr)
		t.Fatalf("Expected array length to be %d but received array of length %d instead\n", len(expected), len(arr))
	}

	for i, s := range arr {
		if s != expected[i] {
			t.Errorf("Expected item at index %d to be %q but received %q instead.\n", i, expected[i], s)
		}
	}
}

func getImportStrings(f FileToken) []string {
	importStrings := make([]string, len(f.Imports))
	i := 0
	for pth := range f.Imports {
		importStrings[i] = pth
		i++
	}
	return importStrings
}
