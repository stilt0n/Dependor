package tokenizer

type FileToken struct {
	FilePath    string
	Imports     map[string][]string
	ReExports   []string
	Exports     []string
	ReExportMap map[string]string
}
