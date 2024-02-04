package dependencygraph

import (
	"dependor/lib/config"
	"dependor/lib/tokenizer"
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"sync"
)

type EdgeList struct {
	edges map[string][]string
	mu    sync.Mutex
}

func (list *EdgeList) Store(key string, value []string) {
	list.mu.Lock()
	defer list.mu.Unlock()
	list.edges[key] = value
}

func (list *EdgeList) Edges() map[string][]string {
	return list.edges
}

type DependencyGraph struct {
	edgeList *EdgeList
	config   *config.Config
	rootPath string
}

func New() *DependencyGraph {
	cfg, err := config.ReadConfig()
	if err != nil {
		// TODO: determine how best to handle / ignore errors
		fmt.Printf("Recieved error but error might be due to using default config settings. See error: %s", err)
	}
	return &DependencyGraph{
		config:   cfg,
		rootPath: ".",
	}
}

// walks file tree starting from current directory and creates an import graph
func (graph *DependencyGraph) Walk() map[string][]string {
	// TODO: make this configurable
	searchableExtensions := regexp.MustCompile("(.js|.jsx|.ts|.tsx)$")
	var wg sync.WaitGroup

	err := filepath.WalkDir(".", func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			fmt.Printf("There was an error accessing path %q: %v\n", path, err)
			return err
		}

		if info.IsDir() && graph.config.ShouldIgnore(path) {
			fmt.Printf("Skipping ignored dir: %+v \n", info.Name())
			return filepath.SkipDir
		}

		if searchableExtensions.MatchString(info.Name()) {
			// TODO: consider limiting number of active threads
			wg.Add(1)
			go graph.readImports(path, &wg)
		}
		return nil
	})

	wg.Wait()

	if err != nil {
		fmt.Printf("Encountered errors while walking file tree: %v\n", err)
	}

	return graph.edgeList.Edges()
}

// I need to look into if sharing resources on a struct is a problem. In that
// case these may need to be refactored into non-recievers
// Note: check that file is .js|.ts|.jsx etc... before calling this so we don't
// waste time reading files that won't have imports
func (graph *DependencyGraph) readImports(filepath string, wg *sync.WaitGroup) {
	defer wg.Done()
	tk, err := tokenizer.NewTokenizerFromFile(filepath)
	if err != nil {
		fmt.Printf("WARN: Failed to read file from path %s.\nReceived error: %s\n Skipping...\n", filepath, err)
	}
	imports := tk.TokenizeImports()
	graph.edgeList.Store(filepath, imports)
}
