package tokenizer

type ImportToken struct {
	ImportPath          string
	ImportedIdentifiers []string
}

type FileToken struct {
	FilePath  string
	Imports   []ImportToken
	ReExports []string
	Exports   []string
}

// Utilties for testing
func (f *FileToken) ImportStrings() []string {
	var importStrings []string
	for _, imp := range f.Imports {
		importStrings = append(importStrings, imp.ImportPath)
	}
	return importStrings
}
