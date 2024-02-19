package dependencygraph

import (
	"dependor/lib/config"
	"dependor/lib/tokenizer"
	"dependor/lib/utils"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
)

type SingleThreadedGraph struct {
	tokens   map[string]*tokenizer.FileToken
	config   *config.Config
	rootPath string
	edgeList map[string][]string
}

// Supports single optional rootPath argument. Uses "." by default.
func NewSync(rootPath ...string) *SingleThreadedGraph {
	rp := "."
	if len(rootPath) > 0 {
		rp = rootPath[0]
	}
	configPath := rp + "/dependor.json"
	if _, err := os.Stat(configPath); rp != "." && err != nil {
		panic(fmt.Sprintf("Root path does not exist or does not have a valid dependor.json config file. To use default config, omit rootPath arg. See error %s\n", err))
	}

	cfg, err := config.ReadConfig(rp + "/dependor.json")
	if err != nil {
		fmt.Println("WARN: No dependor.json file was found so the default config is being used.")
	}

	return &SingleThreadedGraph{
		config:   cfg,
		rootPath: rp,
		tokens:   make(map[string]*tokenizer.FileToken, 0),
	}
}

func (graph *SingleThreadedGraph) ParseGraph() (map[string][]string, error) {
	err := graph.walk()
	if err != nil {
		return nil, err
	}
	graph.resolveImportExtensions()
	graph.finishIndexMaps()
	graph.parseTokens()
	if graph.edgeList == nil {
		return nil, errors.New("parse tokens failed with nil edgeList")
	}
	return graph.edgeList, nil
}

// Walks file tree from root path and populates tokenizedFiles
func (graph *SingleThreadedGraph) walk() error {
	searchableExtensions := regexp.MustCompile(`(\.js|\.jsx|\.ts|\.tsx)$`)
	err := filepath.WalkDir(graph.rootPath, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			fmt.Printf("There was an error accessing path %q: %v\n", path, err)
			return err
		}

		if info.IsDir() && graph.config.ShouldIgnore(path) {
			fmt.Printf("Ignoring directory %q\n", info.Name())
			return filepath.SkipDir
		}

		if searchableExtensions.MatchString(info.Name()) {
			graph.readImports(path)
		}
		return nil
	})

	return err
}

func (graph *SingleThreadedGraph) parseTokens() {
	graph.edgeList = make(map[string][]string, len(graph.tokens))
	for _, tk := range graph.tokens {
		edges := make([]string, 0)
		for importPath, importIdents := range tk.Imports {
			if isIndexFile(importPath) {
				edges = append(edges, graph.resolveIndexImport(importPath, importIdents)...)
			} else {
				edges = append(edges, importPath)
			}
		}
		graph.edgeList[tk.FilePath] = edges
	}
}

func (graph *SingleThreadedGraph) readImports(filePath string) {
	tk, err := tokenizer.NewTokenizerFromFile(filePath)
	if err != nil {
		return
	}
	tokenizedFile := tk.Tokenize()
	graph.tokens[tokenizedFile.FilePath] = &tokenizedFile
}

func (graph *SingleThreadedGraph) resolveImportExtensions() {
	for _, tk := range graph.tokens {
		updatedImports := make(map[string][]string, 0)
		for originalPath, idents := range tk.Imports {
			updatedPath := withExtension(graph.tokens, graph.config, originalPath)
			updatedImports[updatedPath] = idents
		}
		tk.Imports = updatedImports

		if len(tk.ReExports) == 0 {
			continue
		}

		// ReExports aren't needed for withExtension to work so they
		// can be safely overwritten in-place
		for i, originalPath := range tk.ReExports {
			tk.ReExports[i] = withExtension(graph.tokens, graph.config, originalPath)
		}

		for k, v := range tk.ReExportMap {
			// We need to know if a file is referenced in a wildcard export in order
			// to resolve that export. But to check, we will need the file's path to
			// be discoverable in the re-export map.
			if v == "*" {
				tk.ReExportMap[withExtension(graph.tokens, graph.config, k)] = v
				continue
			}
			tk.ReExportMap[k] = withExtension(graph.tokens, graph.config, v)
		}
	}
}

func (graph *SingleThreadedGraph) resolveIndexImport(pth string, idents []string) []string {
	resolvedPaths := make(utils.Set[string], 0)
	for _, ident := range idents {
		if slices.Contains(graph.tokens[pth].Exports, ident) {
			resolvedPaths.Add(pth)
			continue
		}
		resolved, ok := graph.tokens[pth].ReExportMap[ident]
		if !ok {
			continue
		}
		resolvedPaths.Add(resolved)
	}
	return resolvedPaths.Keys()
}

func (graph *SingleThreadedGraph) finishIndexMaps() {
	for _, tk := range graph.tokens {
		// For now I am not supporting re-exports from non-index files but since
		// it seems like most of the work for doing this is finished, I may do
		// so in the future.
		if tk.ReExportMap == nil || !isIndexFile(tk.FilePath) {
			continue
		}

		for _, reExportPath := range tk.ReExports {
			if _, ok := tk.ReExportMap[reExportPath]; !ok {
				continue
			}
			reExportFileNode, ok := graph.tokens[reExportPath]
			if !ok {
				continue
			}
			for _, export := range reExportFileNode.Exports {
				tk.ReExportMap[export] = reExportPath
			}
		}
	}
}

// Resolves any aliases and finds the correct file extension for a path
func withExtension(pathMap map[string]*tokenizer.FileToken, cfg *config.Config, path string) string {
	path = cfg.ReplaceAliases(path)
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

var indexFilePattern = regexp.MustCompile("index.(js|ts|jsx|tsx)$")

func isIndexFile(filePath string) bool {
	return indexFilePattern.MatchString(filePath)
}
