package dependencygraph

import (
	"dependor/lib/config"
	"dependor/lib/tokenizer"
	"fmt"
	"io/fs"
	"os"
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

func newEdgeList() *EdgeList {
	return &EdgeList{
		edges: make(map[string][]string),
	}
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
		fmt.Printf("received error but error might be due to using default config settings. See error: %s\n", err)
	}
	return &DependencyGraph{
		config:   cfg,
		rootPath: ".",
		edgeList: newEdgeList(),
	}
}

func NewWithRootPath(rootPath string) *DependencyGraph {
	if _, err := os.Stat(rootPath); err != nil {
		panic(fmt.Sprintf("Root path does not appear to be a real path. See error:\n  %s\n", err))
	}

	cfg, err := config.ReadConfig(rootPath + "/dependor.json")
	if err != nil {
		// TODO: determine how best to handle / ignore errors
		fmt.Printf("received error but error might be due to using default config settings. See error: %s\n", err)
	}
	return &DependencyGraph{
		config:   cfg,
		rootPath: rootPath,
		edgeList: newEdgeList(),
	}
}

// walks file tree starting from current directory and creates an import graph
func (graph *DependencyGraph) Walk() (map[string][]string, error) {
	// TODO: make this configurable
	searchableExtensions := regexp.MustCompile(`(\.js|\.jsx|\.ts|\.tsx)$`)
	var wg sync.WaitGroup

	err := filepath.WalkDir(graph.rootPath, func(path string, info fs.DirEntry, err error) error {
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
		return make(map[string][]string), err
	}

	graph.resolveImportExtensions()

	return graph.edgeList.Edges(), nil
}

// Note: This exists to make testing easier and will likely be removed in the future
func (graph *DependencyGraph) WalkSync() (map[string][]string, error) {
	searchableExtensions := regexp.MustCompile(`(\.js|\.jsx|\.ts|\.tsx)$`)
	err := filepath.WalkDir(graph.rootPath, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			fmt.Printf("There was an error accessing path %q: %v\n", path, err)
			return err
		}

		if info.IsDir() && graph.config.ShouldIgnore(path) {
			fmt.Printf("Ran into an ignore directory: %q\n", info.Name())
			return filepath.SkipDir
		}

		if searchableExtensions.MatchString(info.Name()) {
			graph.readImportsSync(path)
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Got error: %s", err)
		return make(map[string][]string), err
	}

	graph.resolveImportExtensions()

	return graph.edgeList.Edges(), nil
}

func (graph *DependencyGraph) PrintPaths() {
	searchableExtensions := regexp.MustCompile(`(\.js|\.jsx|\.ts|\.tsx)$`)
	err := filepath.WalkDir(graph.rootPath, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			fmt.Printf("There was an error accessing path %q: %v\n", path, err)
			return err
		}

		if info.IsDir() && graph.config.ShouldIgnore(path) {
			fmt.Printf("Ran into an ignore directory: %q\n", info.Name())
			return filepath.SkipDir
		}

		if searchableExtensions.MatchString(info.Name()) {
			fmt.Printf("Found %q at path %s\n", info.Name(), path)
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Got error: %s", err)
	}
}

func (graph *DependencyGraph) readImports(filepath string, wg *sync.WaitGroup) {
	defer wg.Done()
	tk, err := tokenizer.NewTokenizerFromFile(filepath)
	if err != nil {
		fmt.Printf("WARN: Failed to read file from path %s.\nReceived error: %s\n Skipping...\n", filepath, err)
	}
	imports := tk.TokenizeImports()
	graph.edgeList.Store(filepath, imports)
}

// Note: this exists to make testing easier and will likely be removed in the future
// It will also be useful for benchmarking against the concurrent version to make
// sure that the concurrent version really improves performance
func (graph *DependencyGraph) readImportsSync(filepath string) {
	tk, err := tokenizer.NewTokenizerFromFile(filepath)
	if err != nil {
		fmt.Printf("WARN: Failed to read file from path %s.\nReceived error: %s\n Skipping...\n", filepath, err)
	}
	imports := tk.TokenizeImports()
	graph.edgeList.Store(filepath, imports)
}

// TODO: Optimize
func (graph *DependencyGraph) resolveImportExtensions() {
	edges := graph.edgeList.Edges()
	for file, imports := range edges {
		var updated []string
		for _, path := range imports {
			updated = append(updated, withExtension(edges, path))
		}
		graph.edgeList.Store(file, updated)
	}
}

// TODO: check if this performs better using precompiled regular expression
// TODO: handle cases like `import { x } from '.';`
func withExtension(pathMap map[string][]string, path string) string {
	extensions := []string{
		".js",
		".ts",
		".jsx",
		".tsx",
		"/index.js",
		"/index.ts",
		"/index.jsx",
		"/index.tsx",
	}

	for _, extension := range extensions {
		if _, ok := pathMap[path+extension]; ok {
			return path + extension
		}
	}

	return path
}
