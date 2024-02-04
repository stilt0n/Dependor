package config

import "testing"

func TestReadConfig(t *testing.T) {
	cfg, err := ReadConfig()
	expectedIgnorePatterns := []string{
		"node_modules",
		"*/noRead.js",
	}

	if err != nil {
		t.Fatalf("Got an error when reading config. Error: %s\n", err)
	}

	for i, p := range cfg.IgnorePatterns {
		if p != expectedIgnorePatterns[i] {
			t.Errorf("Recieved an incorrect ignore pattern. Expected: %s. Recieved: %s", expectedIgnorePatterns[i], p)
		}
	}

	if cfg.ShouldIgnore("./okay.js") {
		t.Errorf("Expected ./okay.js not to be ignored")
	}

	if !cfg.ShouldIgnore("node_modules") {
		t.Errorf("Expected node_modules to be ignored")
	}

	if !cfg.ShouldIgnore("dir/noRead.js") {
		t.Errorf("Expected dir/noRead.js to be ignored")
	}
}
