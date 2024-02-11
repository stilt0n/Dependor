package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

const config_file_name = "dependor.json"

type Config struct {
	// These patterns should work with go's `filepath.Match` function, which means no recursive directory mathing.
	// This is a pretty big limitation so I may want to add a glob library like https://github.com/gobwas/glob.
	IgnorePatterns []string `json:"ignorePatterns"`
	// This allows you to resolve paths like `'~/components/Foo'` or `'@monorepo/package/dir/file'`
	PathAliases map[string]string `json:"pathAliases"`
}

func ReadConfig(path ...string) (*Config, error) {
	var readFrom string
	if len(path) > 0 {
		readFrom = path[0]
	} else {
		readFrom = config_file_name
	}
	defaultConfig := &Config{
		IgnorePatterns: []string{"**/node_modules"},
	}
	// By default we assume config is located in the same directory ReadConfig is called from
	// But ReadConfig supports an optional path argument which allows you to read a config
	// from elsewhere.
	configFile, err := os.Open(readFrom)
	if err != nil {
		fmt.Println("Did not find config file in root dir. Using default config.")
		return defaultConfig, err
	}
	defer configFile.Close()

	bytes, err := io.ReadAll(configFile)
	if err != nil {
		fmt.Printf("WARN: ran into unexpected error reading the file. Using default config as a fallback. See error below for more details:\n%s\n", err)
		return defaultConfig, err
	}
	var config Config
	err = json.Unmarshal(bytes, &config)
	if err != nil {
		fmt.Printf("WARN: received an error while parsing defendor.json. Using default config as a fallback. See error below for more details:\n%s\n", err)
		return defaultConfig, err
	}

	return &config, nil
}

func (cfg *Config) ShouldIgnore(path string) bool {
	for _, pattern := range cfg.IgnorePatterns {
		pathMatches, err := doublestar.PathMatch(pattern, path)
		if err != nil {
			panic(fmt.Sprintf("there was an error with the config glob pattern %q. received error: %s", pattern, err))
		}
		if pathMatches {
			return true
		}
	}
	return false
}

// Replaces the first matching alias found in the config or returns the orginal path
// Assumes alias will be at the beginning of the path since that's generally how
// imports are written in JavaScript
func (cfg *Config) ReplaceAliases(path string) string {
	for alias, replacement := range cfg.PathAliases {
		if strings.HasPrefix(path, alias) {
			return strings.Replace(path, alias, replacement, 1)
		}
	}
	return path
}
