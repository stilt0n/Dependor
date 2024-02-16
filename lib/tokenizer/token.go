package tokenizer

type FileToken struct {
	FilePath    string
	Imports     map[string][]string
	ReExports   []string
	Exports     []string
	ReExportMap map[string]string
}

// Utilties for testing
func (f *FileToken) ImportStrings() []string {
	importStrings := make([]string, len(f.Imports))
	i := 0
	for pth := range f.Imports {
		importStrings[i] = pth
		i++
	}
	return importStrings
}
