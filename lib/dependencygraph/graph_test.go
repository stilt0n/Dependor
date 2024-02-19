package dependencygraph

import (
	"slices"
	"testing"
)

func TestParse(t *testing.T) {
	expected_tree := map[string][]string{
		"test_tree/a.js":                                              {"foo", "test_tree/re-exports/rexc.js", "test_tree/re-exports/rexb.js"},
		"test_tree/b.ts":                                              {"foo"},
		"test_tree/util/c.js":                                         {"lodash", "express", "test_tree/b.ts", "test_tree/util/fake_url/printFunc"},
		"test_tree/src/components/d.jsx":                              {"react", "@remix-run/react", "test_tree/src/components/i/i.jsx"},
		"test_tree/src/components/e.tsx":                              {},
		"test_tree/src/components/i/i.jsx":                            {},
		"test_tree/src/components/i/not_imported.ts":                  {"test_tree/re-exports/rexa.js", "test_tree/re-exports/rexb.js"},
		"test_tree/src/components/i/annoying.jsx":                     {"test_tree/src/components/i/i.jsx"},
		"test_tree/src/components/i/folder/importFromParentFolder.ts": {"test_tree/src/components/i/i.jsx"},
		"test_tree/src/components/sibling/importFromSibling.js":       {"test_tree/src/components/i/index.js"},
		"test_tree/src/components/i/index.js":                         {},
		"test_tree/src/components/f.tsx":                              {"react", "dynamic_data"},
		"test_tree/src/hooks/g.ts":                                    {},
		"test_tree/src/hooks/h.ts":                                    {"test_tree/src/hooks/g.ts", "test_tree/a.js", "test_tree/src/hooks/spurious_imports.txt"},
		"test_tree/re-exports/index.js":                               {},
		"test_tree/re-exports/rexa.js":                                {},
		"test_tree/re-exports/rexb.js":                                {},
		"test_tree/re-exports/rexc.js":                                {},
	}

	dgraph := NewSync()

	tree, err := dgraph.ParseGraph()

	if err != nil {
		t.Fatalf("Expected no error. Got: %s\n", err)
	}

	for k, v := range tree {
		expected, ok := expected_tree[k]
		if !ok {
			t.Errorf("Unexpected path in dependency tree. Got %s\n", k)
		}

		if len(expected) != len(v) {
			t.Log(v)
			t.Fatalf("Wrong number of items in adjaceny list for %q. Got %d. Expected %d\n", k, len(v), len(expected))
		}

		// Now that tokenized imports are returned in a map the ordering is non-deterministic
		// This test should be rewritten when dependencygraph is rewritten to use FileTokens
		slices.Sort(expected)
		slices.Sort(v)
		for i, p := range v {
			if expected[i] != p {
				t.Errorf("Expected import path at index %d to be %q. Got %q\n", i, expected[i], p)
			}
		}
	}
}
