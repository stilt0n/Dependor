package config

import (
	"encoding/json"
	"testing"
)

func TestReadConfig(t *testing.T) {
	cfg, err := ReadConfig()
	expectedIgnorePatterns := []string{
		"**/node_modules",
		"*/noRead.js",
	}

	expectedAliases := map[string]string{
		"@monorepo/package": "root/package",
		"~":                 "root/home",
	}

	if err != nil {
		t.Fatalf("got an error when reading config. error: %s\n", err)
	}

	if success := testSliceMatch(t, cfg.IgnorePatterns, expectedIgnorePatterns); !success {
		t.Error("received at least one incorrect ignore pattern\n")
	}

	if success := testMapMatch(t, cfg.PathAliases, expectedAliases); !success {
		t.Error("received at least on incorrect path alias\n")
	}

	type Custom struct {
		FindRelatedTestsOptions struct {
			TestPattern string `json:"testPattern"`
		} `json:"findRelatedTestsOptions"`
	}
	var options Custom

	jsonBytes, err := cfg.GetCustomConfig()
	if err != nil {
		t.Fatalf("Expected no error when marshaling CustomConfig into JSON. Received: %s\n", err)
	}

	if err := json.Unmarshal(jsonBytes, &options); err != nil {
		t.Fatalf("Expected no error when unmarshaling CustomConfig JSON. Received: %s\n", err)
	}

	expected := "(.spec.|.test.)(js|jsx|ts|tsx)$"
	if res := options.FindRelatedTestsOptions.TestPattern; res != expected {
		t.Fatalf("Expected test pattern to be %q but received %q\n", expected, options.FindRelatedTestsOptions.TestPattern)
	}
}

func TestReplacePath(t *testing.T) {
	cfg, err := ReadConfig()
	if err != nil {
		t.Fatalf("got an error when reading config. error: %s\n", err)
	}

	if cfg.ReplaceAliases("@monorepo/package/component/Foo.tsx") != "root/package/component/Foo.tsx" {
		t.Fatalf("Incorrect replacement for '@monorepo/package/component/Foo.tsx'")
	}
}

func TestIgnorePath(t *testing.T) {
	cfg, err := ReadConfig()
	if err != nil {
		t.Fatalf("got an error when reading config. error: %s\n", err)
	}

	if !cfg.ShouldIgnore("node_modules") {
		t.Error("expected node_modules to be ignored")
	}

	if !cfg.ShouldIgnore("base/node_modules") {
		t.Error("expected base/node_modules to be ignored")
	}

	if !cfg.ShouldIgnore("whoo/this/is/pretty/nested/node_modules") {
		t.Error("expected nested node_modules to be ignored")
	}

	if !cfg.ShouldIgnore("base/noRead.js") {
		t.Error("expected base/noRead.js to be ignored")
	}

	if cfg.ShouldIgnore("this/path/is/ok") {
		t.Error("expected this/path/is/ok not to be ignored")
	}
}

func testSliceMatch(t *testing.T, received, expected []string) bool {
	if len(received) != len(expected) {
		t.Errorf("received array of length %d but expected array of length %d\n", len(received), len(expected))
		return false
	}

	success := true
	for index, item := range received {
		if item != expected[index] {
			t.Errorf("expected %q at index %d but received %q instead\n", expected[index], index, item)
			success = false
		}
	}
	return success
}

func testMapMatch(t *testing.T, received, expected map[string]string) bool {
	if len(received) != len(expected) {
		t.Errorf("received map of length %d but expected map of length %d\n", len(received), len(expected))
		return false
	}

	success := true
	for key, val := range received {
		expectedValue, ok := expected[key]
		if !ok {
			t.Errorf("received unexpected key: %q\n", key)
			success = false
			continue
		}
		if val != expectedValue {
			t.Errorf("expected value for key %q to be %q but received %q\n", key, expectedValue, val)
			success = false
		}
	}
	return success
}
