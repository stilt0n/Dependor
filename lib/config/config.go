package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const config_file_name = "dependor.json"

type Config struct {
	// These patterns should work with go's `filepath.Match` function.
	// I may add https://github.com/gobwas/glob or something similar for globs later
	IgnorePatterns []string `json:"ignorePatterns"`
}

func ReadConfig() (*Config, error) {
	defaultConfig := &Config{
		IgnorePatterns: []string{"node_modules"},
	}
	// Right now we always expect the config file to be located int the current
	// working directory but at some point it might make sense to bubble up and
	// look for configs when inside a git repository
	configFile, err := os.Open(config_file_name)
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
		fmt.Printf("WARN: recieved an error while parsing defendor.json. Using default config as a fallback. See error below for more details:\n%s\n", err)
		return defaultConfig, err
	}

	return &config, nil
}

func (c *Config) ShouldIgnore(path string) bool {
	for _, p := range c.IgnorePatterns {
		pathMatches, err := filepath.Match(p, path)
		if err != nil {
			// Panic here is probably not going to be the right choice in the long run but I think it will make finding bugs easier while developing
			panic(fmt.Sprintf("Error using ignore patterns from dependor.json file. There may be a problem with the patterns. \n%s\n", err))
		}
		if pathMatches {
			return true
		}
	}
	return false
}
