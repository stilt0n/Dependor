package config

import "testing"

func TestReadConfig(t *testing.T) {
	cfg, err := ReadConfig()
	expectedIgnorePatterns := []string{
		"node_modules",
		"*/noRead.js",
	}

	expectedAliases := map[string]string{
		"@monorepo/package": "root/package",
		"~":                 "root/home",
	}

	if err != nil {
		t.Fatalf("got an error when reading config. Error: %s\n", err)
	}

	if success := testSliceMatch(t, cfg.IgnorePatterns, expectedIgnorePatterns); !success {
		t.Errorf("recieved at least one incorrect ignore pattern")
	}

	if success := testMapMatch(t, cfg.PathAliases, expectedAliases); !success {
		t.Errorf("recieved at least on incorrect path alias")
	}
}

func testSliceMatch(t *testing.T, recieved, expected []string) bool {
	if len(recieved) != len(expected) {
		t.Errorf("recieved array of length %d but expected array of length %d", len(recieved), len(expected))
		return false
	}

	success := true
	for index, item := range recieved {
		if item != expected[index] {
			t.Errorf("expected %q at index %d but recieved %q instead", expected[index], index, item)
			success = false
		}
	}
	return success
}

func testMapMatch(t *testing.T, recieved, expected map[string]string) bool {
	if len(recieved) != len(expected) {
		t.Errorf("recieved map of length %d but expected map of length %d", len(recieved), len(expected))
		return false
	}

	success := true
	for key, val := range recieved {
		expectedValue, ok := expected[key]
		if !ok {
			t.Errorf("recieved unexpected key: %q", key)
			success = false
			continue
		}
		if val != expectedValue {
			t.Errorf("expected value for key %q to be %q but recieved %q", key, expectedValue, val)
			success = false
		}
	}
	return success
}
